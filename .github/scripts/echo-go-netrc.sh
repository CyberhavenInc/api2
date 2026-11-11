# Sourced (not executed) by workflow steps that need golang.echohq.com.
# Writes a short-lived netrc with the Echo Libraries key, points the go
# command at it, and removes it when the sourcing step's shell exits.
if [ -z "${ECHO_LIBRARIES_USERNAME:-}" ] || [ -z "${ECHO_LIBRARIES_KEY:-}" ]; then
  echo "::error::ECHO_LIBRARIES_USERNAME and ECHO_LIBRARIES_KEY secrets are required for golang.echohq.com downloads"
  exit 1
fi
ECHO_NETRC_DIR="${RUNNER_TEMP:-/tmp}"
if command -v cygpath >/dev/null 2>&1; then
  ECHO_NETRC_DIR="$(cygpath -u "$ECHO_NETRC_DIR")"
fi
ECHO_NETRC_FILE="$(mktemp "$ECHO_NETRC_DIR/echo-go-netrc.XXXXXX")"
trap 'rm -f "$ECHO_NETRC_FILE"' EXIT
chmod 600 "$ECHO_NETRC_FILE"
printf 'machine golang.echohq.com\n  login %s\n  password %s\n' \
  "$ECHO_LIBRARIES_USERNAME" "$ECHO_LIBRARIES_KEY" > "$ECHO_NETRC_FILE"
# On Windows runners bash is Git Bash, but go.exe needs a native path.
if command -v cygpath >/dev/null 2>&1; then
  NETRC="$(cygpath -w "$ECHO_NETRC_FILE")"
else
  NETRC="$ECHO_NETRC_FILE"
fi
export NETRC
export GOPROXY=https://golang.echohq.com
