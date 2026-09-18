# https://just.systems

test:
    gotestsum --format short ./...
    
test-watch:
    gotestsum --watch --format short ./...
