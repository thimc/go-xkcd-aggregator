NAME = goxkcd

$(NAME):
	@mkdir bin
	@go build -o ./bin/$(NAME) ./cmd/$(NAME)/$(NAME).go

clean:
	rm -f $(NAME)

.PHONY: $(NAME) clean
