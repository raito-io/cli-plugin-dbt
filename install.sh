# This tool is meant for development reasons.
# It builds the plugin and installs it in the global plugins folder of the CLI
go mod tidy
make wire
go build -o  ~/.raito/plugins/raito-io/cli-plugin-dbt-latest .
