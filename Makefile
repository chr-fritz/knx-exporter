.PHONY: completions
completions:
	rm -rf completions
	mkdir completions
	for sh in bash zsh; do go run main.go completion "$$sh" >"completions/knx-exporter.$$sh"; done