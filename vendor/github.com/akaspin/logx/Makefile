REPO	= github.com/akaspin/logx

TESTS	      = .
TEST_TAGS     =
TEST_ARGS     =
BENCH	      = .

test:
	go test -v -race -tags="trace"
	go test -v -race -tags="debug"
	go test -v -race
	go test -v -race -tags="notice"
