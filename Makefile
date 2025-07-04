run:
	@go run cmd/api/main.go

draw:
	@dot -Tsvg -O my-graph.gv
