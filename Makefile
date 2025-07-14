run:
	@go run cmd/api/main.go

draw:
	@for file in *.gv; do \
		dot -Tsvg "$$file" -O; \
	done
