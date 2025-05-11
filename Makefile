draw:
	@go run *.go
	@dot -Tsvg -O my-graph.gv
