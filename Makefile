LOCALSTACK_CONTAINER=fkcp-localstack
SNS_TOPIC_NAME=TESTE
SNS_ENDPOINT=http://localhost:4566
AWS_REGION=us-east-1

.PHONY: run
run:
	@go run cmd/api/main.go

.PHONY: draw
draw:
	@for file in *.gv; do \
		dot -Tsvg "$$file" -O; \
	done

.PHONY: localstack
localstack: start-localstack create-sns-topic

.PHONY: start-localstack
start-localstack:
	@if ! docker ps --format '{{.Names}}' | grep -q '^$(LOCALSTACK_CONTAINER)$$'; then \
		docker start $(LOCALSTACK_CONTAINER); \
		sleep 5; \
	fi

.PHONY: create-sns-topic
create-sns-topic:
	@aws --endpoint-url=$(SNS_ENDPOINT) sns create-topic --name $(SNS_TOPIC_NAME) --region $(AWS_REGION)
