draw:
	@go run cmd/api/main.go
	@dot -Tsvg -O my-graph.gv
