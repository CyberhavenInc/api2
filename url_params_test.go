package api2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifier(t *testing.T) {
	cl := newPathClassifier([]string{
		"/users",
		"/users/:user",
		"/users/:user/posts",
		"/users/:user/posts/:post",
		"/users/:user/posts/:post/comments",
		"/users/:user/comments",
		"/users/:user/comments/:comment",
		"/users/:user/comments/:comment/responses",
	})

	cases := []struct {
		url             string
		wantIndex       int
		wantParam2value map[string]string
	}{
		{
			url:             "/users",
			wantIndex:       0,
			wantParam2value: map[string]string{},
		},
		{
			url:       "/user",
			wantIndex: -1,
		},
		{
			url:       "/users2",
			wantIndex: -1,
		},
		{
			url:       "/users/123",
			wantIndex: 1,
			wantParam2value: map[string]string{
				"user": "123",
			},
		},
		{
			url:       "/users/123/",
			wantIndex: 1,
			wantParam2value: map[string]string{
				"user": "123",
			},
		},
		{
			url:       "/users/123/test",
			wantIndex: -1,
		},
		{
			url:       "/users/123/posts",
			wantIndex: 2,
			wantParam2value: map[string]string{
				"user": "123",
			},
		},
		{
			url:       "/users/123/posts/456-789",
			wantIndex: 3,
			wantParam2value: map[string]string{
				"user": "123",
				"post": "456-789",
			},
		},
		{
			url:       "/users/123/posts/456-789/",
			wantIndex: 3,
			wantParam2value: map[string]string{
				"user": "123",
				"post": "456-789",
			},
		},
		{
			url:       "/users/123/posts/456-789/comments",
			wantIndex: 4,
			wantParam2value: map[string]string{
				"user": "123",
				"post": "456-789",
			},
		},
		{
			url:       "/users/123/posts/456-789/comments/",
			wantIndex: 4,
			wantParam2value: map[string]string{
				"user": "123",
				"post": "456-789",
			},
		},
		{
			url:       "/users/123/comments",
			wantIndex: 5,
			wantParam2value: map[string]string{
				"user": "123",
			},
		},
		{
			url:       "/users/123/comments/",
			wantIndex: 5,
			wantParam2value: map[string]string{
				"user": "123",
			},
		},
		{
			url:       "/users/123/comments/ab505",
			wantIndex: 6,
			wantParam2value: map[string]string{
				"user":    "123",
				"comment": "ab505",
			},
		},
		{
			url:       "/users/123/comments/ab505/",
			wantIndex: 6,
			wantParam2value: map[string]string{
				"user":    "123",
				"comment": "ab505",
			},
		},
		{
			url:       "/users/123/comments/ab505/responses",
			wantIndex: 7,
			wantParam2value: map[string]string{
				"user":    "123",
				"comment": "ab505",
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.url, func(t *testing.T) {
			gotIndex, gotParam2value := cl.Classify(tc.url)
			require.Equal(t, tc.wantIndex, gotIndex)
			require.Equal(t, tc.wantParam2value, gotParam2value)
		})
	}
}

func TestClassifierWithWildcards(t *testing.T) {
	cl := newPathClassifier([]string{
		"/data/:path*",
		"/files/:name/content",
		"/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part",
		"/api/:version/items/:id*",
		"/greedy/:first*/middle/:second*/end",
	})

	cases := []struct {
		name            string
		url             string
		wantIndex       int
		wantParam2value map[string]string
	}{
		{
			name:      "simple wildcard with multiple segments",
			url:       "/data/a/b/c/d",
			wantIndex: 0,
			wantParam2value: map[string]string{
				"path": "a/b/c/d",
			},
		},
		{
			name:      "wildcard with single segment",
			url:       "/data/single",
			wantIndex: 0,
			wantParam2value: map[string]string{
				"path": "single",
			},
		},
		{
			name:      "wildcard with no segments (greedy)",
			url:       "/data",
			wantIndex: 0,
			wantParam2value: map[string]string{
				"path": "",
			},
		},
		{
			name:      "complex pattern with wildcard in middle",
			url:       "/wildcard/a/static-part/b/b1/b2/b3/b4/static-part/c/static_part",
			wantIndex: 2,
			wantParam2value: map[string]string{
				"param_a": "a",
				"param_b": "b/b1/b2/b3/b4",
				"param_c": "c",
			},
		},
		{
			name:      "complex pattern with minimal wildcard content",
			url:       "/wildcard/x/static-part/y/static-part/z/static_part",
			wantIndex: 2,
			wantParam2value: map[string]string{
				"param_a": "x",
				"param_b": "y",
				"param_c": "z",
			},
		},
		{
			name:      "api with wildcard at end",
			url:       "/api/v1/items/user/123/settings",
			wantIndex: 3,
			wantParam2value: map[string]string{
				"version": "v1",
				"id":      "user/123/settings",
			},
		},
		{
			name:      "api with empty wildcard at end",
			url:       "/api/v2/items",
			wantIndex: 3,
			wantParam2value: map[string]string{
				"version": "v2",
				"id":      "",
			},
		},
		{
			name:      "multiple wildcards",
			url:       "/greedy/first/part/middle/second/part/end",
			wantIndex: 4,
			wantParam2value: map[string]string{
				"first":  "first/part",
				"second": "second/part",
			},
		},
		{
			name:      "non-matching case - should prefer static route",
			url:       "/files/document.txt/content",
			wantIndex: 1,
			wantParam2value: map[string]string{
				"name": "document.txt",
			},
		},
		{
			name:      "non-matching - insufficient static parts",
			url:       "/wildcard/a/static-part/b/b1/wrong-static/c/static_part",
			wantIndex: -1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotIndex, gotParam2value := cl.Classify(tc.url)
			require.Equal(t, tc.wantIndex, gotIndex, "Expected index %d, got %d for URL %s", tc.wantIndex, gotIndex, tc.url)
			require.Equal(t, tc.wantParam2value, gotParam2value, "Parameter mismatch for URL %s", tc.url)
		})
	}
}

func TestRouteSpecificity(t *testing.T) {
	cases := []struct {
		name             string
		routePath        string
		expectedStatic   int
		expectedRegular  int
		expectedWildcard int
		expectedEndsWith bool
		expectedScore    int
	}{
		{
			name:             "simple static route",
			routePath:        "/users",
			expectedStatic:   1,
			expectedRegular:  0,
			expectedWildcard: 0,
			expectedEndsWith: false,
			expectedScore:    100, // 1*100 + 0*10 - 0*50 - 0*25
		},
		{
			name:             "route with regular parameter",
			routePath:        "/users/:id",
			expectedStatic:   1,
			expectedRegular:  1,
			expectedWildcard: 0,
			expectedEndsWith: false,
			expectedScore:    110, // 1*100 + 1*10 - 0*50 - 0*25
		},
		{
			name:             "route with wildcard at end",
			routePath:        "/data/:path*",
			expectedStatic:   1,
			expectedRegular:  0,
			expectedWildcard: 1,
			expectedEndsWith: true,
			expectedScore:    25, // 1*100 + 0*10 - 1*50 - 1*25
		},
		{
			name:             "complex route with multiple parameters",
			routePath:        "/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part",
			expectedStatic:   4, // "wildcard", "static-part", "static-part", "static_part"
			expectedRegular:  2, // param_a, param_c
			expectedWildcard: 1, // param_b*
			expectedEndsWith: false,
			expectedScore:    370, // 4*100 + 2*10 - 1*50 - 0*25
		},
		{
			name:             "very complex route",
			routePath:        "/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part/:param_d*/static_part/:param_e/static_part/:param_f*/static_part",
			expectedStatic:   7, // "wildcard", "static-part", "static-part", "static_part", "static_part", "static_part", "static_part"
			expectedRegular:  3, // param_a, param_c, param_e
			expectedWildcard: 3, // param_b*, param_d*, param_f*
			expectedEndsWith: false,
			expectedScore:    580, // 7*100 + 3*10 - 3*50 - 0*25 = 700 + 30 - 150 = 580
		},
		{
			name:             "simple route with wildcard",
			routePath:        "/wildcard/:param_a/static-part/:param_b*",
			expectedStatic:   2, // "wildcard", "static-part"
			expectedRegular:  1, // param_a
			expectedWildcard: 1, // param_b*
			expectedEndsWith: true,
			expectedScore:    135, // 2*100 + 1*10 - 1*50 - 1*25
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			spec := calculateRouteSpecificity(tc.routePath)
			require.Equal(t, tc.expectedStatic, spec.staticSegments, "Static segments mismatch")
			require.Equal(t, tc.expectedRegular, spec.regularParams, "Regular parameters mismatch")
			require.Equal(t, tc.expectedWildcard, spec.wildcardParams, "Wildcard parameters mismatch")
			require.Equal(t, tc.expectedEndsWith, spec.endsWithWildcard, "Ends with wildcard mismatch")
			require.Equal(t, tc.expectedScore, spec.score, "Specificity score mismatch")
		})
	}
}

func TestRouteConflictDetection(t *testing.T) {
	cases := []struct {
		name              string
		routes            []string
		expectedConflicts int
	}{
		{
			name:              "no conflicts",
			routes:            []string{"GET: /users", "GET: /posts", "GET: /comments"},
			expectedConflicts: 0,
		},
		{
			name:              "simple conflict - wildcard before specific",
			routes:            []string{"GET: /data/:path*", "GET: /data/specific"},
			expectedConflicts: 1,
		},
		{
			name: "wildcard conflict - BasicWildcard vs AdvancedWildcard",
			routes: []string{
				"GET: /wildcard/:param_a/static-part/:param_b*",
				"GET: /wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part/:param_d*/static_part/:param_e/static_part/:param_f*/static_part",
			},
			expectedConflicts: 1,
		},
		{
			name: "multiple conflicts",
			routes: []string{
				"GET: /api/:version*",
				"GET: /api/:version/users/:id*",
				"GET: /api/:version/users/:id/posts",
			},
			expectedConflicts: 3, // route 0 conflicts with routes 1 and 2, route 1 conflicts with route 2
		},
		{
			name: "no conflict - different starting paths",
			routes: []string{
				"GET: /users/:id*",
				"GET: /posts/:id*",
			},
			expectedConflicts: 0,
		},
		{
			name: "multiple conflicts, different methods",
			routes: []string{
				"GET: /api/:version*",
				"GET: /api/:version/users/:id*",
				"POST: /api/:version/users/:id/posts",
			},
			expectedConflicts: 1, // route 0 conflicts with routes 1 but not with 2
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			conflicts := detectRouteConflicts(tc.routes)
			require.Equal(t, tc.expectedConflicts, len(conflicts), "Number of conflicts mismatch")
		})
	}
}

func TestRouteSorting(t *testing.T) {
	cases := []struct {
		name          string
		routes        []string
		expectedOrder []string
	}{
		{
			name: "sort by specificity",
			routes: []string{
				"/data/:path*",       // score: 25
				"/data/specific",     // score: 200
				"/data/:id/comments", // score: 210
			},
			expectedOrder: []string{
				"/data/:id/comments", // score: 210 (highest)
				"/data/specific",     // score: 200
				"/data/:path*",       // score: 25 (lowest)
			},
		},
		{
			name: "AdvancedWildcard/BasicWildcard ordering",
			routes: []string{
				"/wildcard/:param_a/static-part/:param_b*", // score: 135
				"/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part/:param_d*/static_part/:param_e/static_part/:param_f*/static_part", // score: 470
			},
			expectedOrder: []string{
				"/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part/:param_d*/static_part/:param_e/static_part/:param_f*/static_part", // score: 470 (higher)
				"/wildcard/:param_a/static-part/:param_b*", // score: 135 (lower)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sorted := sortRoutesBySpecificity(tc.routes)
			require.Equal(t, tc.expectedOrder, sorted, "Route sorting order mismatch")
		})
	}
}

func TestRoutesCanConflict(t *testing.T) {
	cases := []struct {
		name        string
		route1      string
		route2      string
		canConflict bool
	}{
		{
			name:        "identical routes",
			route1:      "GET: /users/:id",
			route2:      "GET: /users/:id",
			canConflict: true,
		},
		{
			name:        "different static paths",
			route1:      "GET: /users/:id",
			route2:      "GET: /posts/:id",
			canConflict: false,
		},
		{
			name:        "wildcard can conflict with specific",
			route1:      "GET: /data/:path*",
			route2:      "GET: /data/specific/path",
			canConflict: true,
		},
		{
			name:        "wildcard vs simple with same prefix",
			route1:      "GET: /wildcard/:param_a/static-part/:param_b*",
			route2:      "GET: /wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part",
			canConflict: true,
		},
		{
			name:        "no conflict - different lengths without wildcards",
			route1:      "GET: /users/:id",
			route2:      "GET: /users/:id/posts",
			canConflict: false,
		},
		{
			name:        "identical routes, different methods",
			route1:      "GET: /users/:id",
			route2:      "POST: /users/:id",
			canConflict: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := routesCanConflict(tc.route1, tc.route2)
			require.Equal(t, tc.canConflict, result, "Conflict detection result mismatch")
		})
	}
}

func TestSplitUrl(t *testing.T) {
	cases := []struct {
		url  string
		want []string
	}{
		{
			url:  "/foo",
			want: []string{"foo"},
		},
		{
			url:  "//foo",
			want: []string{"foo"},
		},
		{
			url:  "/foo/bar",
			want: []string{"foo", "bar"},
		},
		{
			url:  "/foo/bar/",
			want: []string{"foo", "bar"},
		},
		{
			url:  "/foo/bar//",
			want: []string{"foo", "bar"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.url, func(t *testing.T) {
			got := splitUrl(tc.url)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFindUrlKeys(t *testing.T) {
	cases := []struct {
		mask string
		want []string
	}{
		{
			mask: "/",
			want: []string{},
		},
		{
			mask: "/foo",
			want: []string{},
		},
		{
			mask: "/:foo",
			want: []string{"foo"},
		},
		{
			mask: "/:foo/",
			want: []string{"foo"},
		},
		{
			mask: "/bar/:foo/zoo/:baz",
			want: []string{"foo", "baz"},
		},
		{
			mask: "/data/:path*",
			want: []string{"path"},
		},
		{
			mask: "/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part",
			want: []string{"param_a", "param_b", "param_c"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.mask, func(t *testing.T) {
			got := findUrlKeys(tc.mask)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCutUrlParams(t *testing.T) {
	cases := []struct {
		mask string
		want string
	}{
		{
			mask: "/",
			want: "/",
		},
		{
			mask: "/foo",
			want: "/foo",
		},
		{
			mask: "/bar/:foo",
			want: "/bar/",
		},
		{
			mask: "/bar/:foo/",
			want: "/bar/",
		},
		{
			mask: "/:foo",
			want: "/",
		},
		{
			mask: "/:foo/",
			want: "/",
		},
		{
			mask: "/bar/:foo/zoo/:baz",
			want: "/bar/",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.mask, func(t *testing.T) {
			got := cutUrlParams(tc.mask)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestBuildUrl(t *testing.T) {
	cases := []struct {
		name        string
		mask        string
		param2value map[string]string
		want        string
		wantError   bool
	}{
		{
			name:        "no parameters",
			mask:        "/path",
			param2value: map[string]string{},
			want:        "/path",
		},
		{
			name: "single parameter",
			mask: "/path/:id",
			param2value: map[string]string{
				"id": "123",
			},
			want: "/path/123",
		},
		{
			name: "two parameters",
			mask: "/path/:id/comment/:comment",
			param2value: map[string]string{
				"id":      "123",
				"comment": "444",
			},
			want: "/path/123/comment/444",
		},
		{
			name: "missing parameter in param2value",
			mask: "/path/:id/comment/:comment",
			param2value: map[string]string{
				"comment": "444",
			},
			wantError: true,
		},
		{
			name: "extra parameter in param2value",
			mask: "/path/:id/comment/:comment",
			param2value: map[string]string{
				"id":      "123",
				"comment": "444",
				"user":    "555",
			},
			wantError: true,
		},
		{
			name: "wildcard parameter",
			mask: "/data/:path*",
			param2value: map[string]string{
				"path": "a/b/c/d",
			},
			want: "/data/a/b/c/d",
		},
		{
			name: "complex wildcard pattern",
			mask: "/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part",
			param2value: map[string]string{
				"param_a": "a",
				"param_b": "b/b1/b2/b3/b4",
				"param_c": "c",
			},
			want: "/wildcard/a/static-part/b/b1/b2/b3/b4/static-part/c/static_part",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildUrl(tc.mask, tc.param2value)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
