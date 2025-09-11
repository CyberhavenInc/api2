package api2

import (
	"fmt"
	"strings"
)

type paramMapType struct{}

type wildcardInfo struct {
	position int    // position in the path parts array
	name     string // parameter name without prefix
}

type routeSpecificity struct {
	staticSegments   int  // number of static path segments
	regularParams    int  // number of regular parameters (:param)
	wildcardParams   int  // number of wildcard parameters (:param*)
	endsWithWildcard bool // true if route ends with a wildcard parameter
	score            int  // calculated specificity score
}

type routeConflict struct {
	route1Index int
	route2Index int
	route1Path  string
	route2Path  string
	route1Score int
	route2Score int
	message     string
}

type classifier struct {
	maskPartsArray [][]string
	paramsArray    [][]bool
	paramsNum      []int
	wildcardArray  [][]wildcardInfo // wildcard information for each route
}

func splitUrl(url string) []string {
	parts := strings.Split(url, "/")
	for len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func newPathClassifier(masks []string) *classifier {
	maskPartsArray := make([][]string, 0, len(masks))
	paramsArray := make([][]bool, 0, len(masks))
	paramsNum := make([]int, 0, len(masks))
	wildcardArray := make([][]wildcardInfo, 0, len(masks))

	for _, mask := range masks {
		parts := splitUrl(mask)
		params := make([]bool, len(parts))
		wildcards := make([]wildcardInfo, 0)
		count := 0

		for i, part := range parts {
			if strings.HasPrefix(part, ":") {
				paramName := strings.TrimPrefix(part, ":")

				// Check for wildcard suffix
				if strings.HasSuffix(paramName, "*") {
					// Wildcard parameter (zero or more segments)
					paramName = strings.TrimSuffix(paramName, "*")
					wildcards = append(wildcards, wildcardInfo{
						position: i,
						name:     paramName,
					})
				}

				parts[i] = paramName
				params[i] = true
				count++
			}
		}

		maskPartsArray = append(maskPartsArray, parts)
		paramsArray = append(paramsArray, params)
		paramsNum = append(paramsNum, count)
		wildcardArray = append(wildcardArray, wildcards)
	}

	return &classifier{
		maskPartsArray: maskPartsArray,
		paramsArray:    paramsArray,
		paramsNum:      paramsNum,
		wildcardArray:  wildcardArray,
	}
}

func match(pathParts, maskParts []string, params []bool, count int) (bool, map[string]string) {
	return matchWithWildcards(pathParts, maskParts, params, count, nil)
}

func matchWithWildcards(pathParts, maskParts []string, params []bool, count int, wildcards []wildcardInfo) (bool, map[string]string) {
	// If no wildcards, use the original simple matching logic
	if len(wildcards) == 0 {
		if len(pathParts) != len(maskParts) {
			return false, nil
		}
		// Check if all static parts match.
		for i := 0; i < len(pathParts); i++ {
			if !params[i] && pathParts[i] != maskParts[i] {
				return false, nil
			}
		}
		// Fill values of the parameters.
		param2value := make(map[string]string, count)
		for i := 0; i < len(pathParts); i++ {
			if !params[i] {
				continue
			}
			key := maskParts[i]
			value := pathParts[i]
			param2value[key] = value
		}
		return true, param2value
	}

	// Complex matching with wildcards
	return matchWildcardPattern(pathParts, maskParts, params, wildcards)
}

func matchWildcardPattern(pathParts, maskParts []string, params []bool, wildcards []wildcardInfo) (bool, map[string]string) {
	param2value := make(map[string]string)

	// For each wildcard, we need to determine what segments it should capture
	pathIndex := 0
	maskIndex := 0

	for maskIndex < len(maskParts) {
		// Check if current mask position is a wildcard
		isWildcard := false
		var wildcardInfo wildcardInfo
		for _, wc := range wildcards {
			if wc.position == maskIndex {
				isWildcard = true
				wildcardInfo = wc
				break
			}
		}

		if isWildcard {
			// Find the next static segment after this wildcard to determine boundaries
			nextStaticIndex := findNextStaticSegment(maskParts, maskIndex+1, params)

			if nextStaticIndex == -1 {
				// Wildcard is at the end - capture remaining path segments
				if pathIndex < len(pathParts) {
					remainingSegments := pathParts[pathIndex:]
					param2value[wildcardInfo.name] = strings.Join(remainingSegments, "/")
					pathIndex = len(pathParts) // consumed all remaining segments
				} else {
					// Wildcard at the end can match zero segments
					param2value[wildcardInfo.name] = ""
				}
				// Since wildcard is at the end, we can finish processing
				maskIndex = len(maskParts)
				break
			} else {
				// Find where the next static segment pattern matches in the path
				matchPos := findStaticPattern(pathParts, pathIndex, maskParts, nextStaticIndex, params)
				if matchPos == -1 {
					return false, nil
				}

				// Calculate how many segments the wildcard should capture
				capturedSegments := pathParts[pathIndex:matchPos]

				param2value[wildcardInfo.name] = strings.Join(capturedSegments, "/")
				pathIndex = matchPos
			}
		} else if params[maskIndex] {
			// Regular parameter - must match exactly one segment
			if pathIndex >= len(pathParts) {
				return false, nil
			}
			param2value[maskParts[maskIndex]] = pathParts[pathIndex]
			pathIndex++
		} else {
			// Static segment - must match exactly
			if pathIndex >= len(pathParts) || pathParts[pathIndex] != maskParts[maskIndex] {
				return false, nil
			}
			pathIndex++
		}

		maskIndex++
	}

	// Ensure we've consumed all path segments
	if pathIndex != len(pathParts) {
		return false, nil
	}

	return true, param2value
}

func findNextStaticSegment(maskParts []string, startIndex int, params []bool) int {
	for i := startIndex; i < len(maskParts); i++ {
		if !params[i] {
			return i
		}
	}
	return -1
}

func findStaticPattern(pathParts []string, startIndex int, maskParts []string, maskStartIndex int, params []bool) int {
	// Find the end of the static pattern we're looking for
	patternEnd := maskStartIndex
	for patternEnd < len(maskParts) && !params[patternEnd] {
		patternEnd++
	}

	patternLength := patternEnd - maskStartIndex

	// Search for this pattern in the remaining path
	for i := startIndex; i <= len(pathParts)-patternLength; i++ {
		match := true
		for j := 0; j < patternLength; j++ {
			if pathParts[i+j] != maskParts[maskStartIndex+j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}

	return -1
}

// Classify returns index of matching mask (-1 if not found) and parameters map.
func (c *classifier) Classify(path string) (index int, param2value map[string]string) {
	pathParts := splitUrl(path)
	for i, maskParts := range c.maskPartsArray {
		ok, param2value := matchWithWildcards(pathParts, maskParts, c.paramsArray[i], c.paramsNum[i], c.wildcardArray[i])
		if ok {
			return i, param2value
		}
	}
	return -1, nil
}

func findUrlKeys(mask string) []string {
	parts := strings.Split(mask, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			// Strip wildcard suffixes for parameter name consistency
			if strings.HasSuffix(paramName, "*") {
				paramName = strings.TrimSuffix(paramName, "*")
			}
			result = append(result, paramName)
		}
	}
	return result
}

func cutUrlParams(mask string) string {
	before, _, _ := strings.Cut(mask, "/:")
	if before == mask {
		return mask
	}
	if !strings.HasSuffix(before, "/") {
		before += "/"
	}
	return before
}

func buildUrl(mask string, param2value map[string]string) (string, error) {
	urlParts := strings.Split(mask, "/")
	replaced := make(map[string]struct{}, len(param2value))
	for i, part := range urlParts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		paramName := strings.TrimPrefix(part, ":")
		// Strip wildcard suffixes for parameter lookup
		if strings.HasSuffix(paramName, "*") {
			paramName = strings.TrimSuffix(paramName, "*")
		}
		value, has := param2value[paramName]
		if !has {
			return "", fmt.Errorf("unknown parameter: %s", paramName)
		}
		urlParts[i] = value
		replaced[paramName] = struct{}{}
	}
	if len(replaced) != len(param2value) {
		return "", fmt.Errorf("not all parameters were built into URL: want %d, got %d", len(param2value), len(replaced))
	}
	return strings.Join(urlParts, "/"), nil
}

// calculateRouteSpecificity analyzes a route pattern and returns its specificity metrics
func calculateRouteSpecificity(routePath string) routeSpecificity {
	parts := splitUrl(routePath)

	var staticSegments, regularParams, wildcardParams int
	var endsWithWildcard bool

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			if strings.HasSuffix(paramName, "*") {
				wildcardParams++
				if i == len(parts)-1 {
					endsWithWildcard = true
				}
			} else {
				regularParams++
			}
		} else {
			staticSegments++
		}
	}

	// Specificity scoring: higher score = more specific route
	// Formula: (static_segments × 100) + (regular_params × 10) - (wildcard_params × 50) - (ends_with_wildcard × 25)
	score := (staticSegments * 100) + (regularParams * 10) - (wildcardParams * 50)
	if endsWithWildcard {
		score -= 25
	}

	return routeSpecificity{
		staticSegments:   staticSegments,
		regularParams:    regularParams,
		wildcardParams:   wildcardParams,
		endsWithWildcard: endsWithWildcard,
		score:            score,
	}
}

// detectRouteConflicts analyzes a list of route patterns for potential conflicts
func detectRouteConflicts(routePaths []string) []routeConflict {
	conflicts := make([]routeConflict, 0)

	// Calculate specificity for all routes
	specificities := make([]routeSpecificity, len(routePaths))
	for i, path := range routePaths {
		specificities[i] = calculateRouteSpecificity(strings.Split(path, ": ")[1]) // Ignore HTTP method prefix
	}

	// Check each pair of routes for conflicts
	for i := 0; i < len(routePaths); i++ {
		for j := i + 1; j < len(routePaths); j++ {
			if routesCanConflict(routePaths[i], routePaths[j]) {
				// Determine which route has higher specificity
				route1Spec := specificities[i]
				route2Spec := specificities[j]

				// If a less specific route appears before a more specific one, it's a conflict
				if route1Spec.score < route2Spec.score {
					conflicts = append(conflicts, routeConflict{
						route1Index: i,
						route2Index: j,
						route1Path:  routePaths[i],
						route2Path:  routePaths[j],
						route1Score: route1Spec.score,
						route2Score: route2Spec.score,
						message:     fmt.Sprintf("Route %d (score: %d) is less specific than route %d (score: %d) but appears first", i, route1Spec.score, j, route2Spec.score),
					})
				}
			}
		}
	}

	return conflicts
}

// routesCanConflict checks if two route patterns could potentially match the same URL
func routesCanConflict(route1, route2 string) bool {
	method1 := strings.SplitN(route1, ": ", 2)[0]
	method2 := strings.SplitN(route2, ": ", 2)[0]
	if method1 != method2 {
		return false // Different HTTP methods can't conflict
	}
	parts1 := splitUrl(route1)
	parts2 := splitUrl(route2)

	// Routes can only conflict if they start with the same pattern
	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		part1 := parts1[i]
		part2 := parts2[i]

		// If both are static and different, they can't conflict
		if !strings.HasPrefix(part1, ":") && !strings.HasPrefix(part2, ":") {
			if part1 != part2 {
				return false
			}
			continue
		}

		// If one is a wildcard, they can potentially conflict
		if (strings.HasPrefix(part1, ":") && strings.HasSuffix(part1, "*")) ||
			(strings.HasPrefix(part2, ":") && strings.HasSuffix(part2, "*")) {
			return true
		}

		// Both are parameters (regular or wildcard), they can potentially conflict
		if strings.HasPrefix(part1, ":") && strings.HasPrefix(part2, ":") {
			continue
		}
	}

	// If we've reached here and routes have the same length, they can conflict
	if len(parts1) == len(parts2) {
		return true
	}

	// If one route is longer, they could conflict if the shorter has wildcards
	shorterParts := parts1
	if len(parts2) < len(parts1) {
		shorterParts = parts2
	}

	// Check if the shorter route ends with a wildcard
	if len(shorterParts) > 0 {
		lastPart := shorterParts[len(shorterParts)-1]
		if strings.HasPrefix(lastPart, ":") && strings.HasSuffix(lastPart, "*") {
			return true
		}
	}

	return false
}

// sortRoutesBySpecificity sorts route patterns by their specificity scores (highest first)
func sortRoutesBySpecificity(routePaths []string) []string {
	// Create pairs of (route, specificity) for sorting
	type routeWithSpec struct {
		path          string
		specificity   routeSpecificity
		originalIndex int
	}

	routes := make([]routeWithSpec, len(routePaths))
	for i, path := range routePaths {
		routes[i] = routeWithSpec{
			path:          path,
			specificity:   calculateRouteSpecificity(path),
			originalIndex: i,
		}
	}

	// Sort by specificity score (highest first), then by original index for stability
	for i := 0; i < len(routes); i++ {
		for j := i + 1; j < len(routes); j++ {
			// Higher score should come first
			if routes[i].specificity.score < routes[j].specificity.score ||
				(routes[i].specificity.score == routes[j].specificity.score && routes[i].originalIndex > routes[j].originalIndex) {
				routes[i], routes[j] = routes[j], routes[i]
			}
		}
	}

	// Extract sorted paths
	sortedPaths := make([]string, len(routes))
	for i, route := range routes {
		sortedPaths[i] = route.path
	}

	return sortedPaths
}
