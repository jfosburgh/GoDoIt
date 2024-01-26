run:
	go run main.go

css:
	tailwindcss -i css/input.css -o css/output.css --minify
