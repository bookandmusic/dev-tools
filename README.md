# Dev-Tools Plugin System

## Plugin System

Dev-Tools supports three types of plugins:

1. **Shell Plugins**: Execute shell scripts with argument passing
2. **Ansible Plugins**: Run Ansible playbooks
3. **Go Plugins**: Internal plugins written in Go

### Plugin Metadata

Each external plugin must have a [meta.yml](plugins/shell/example/meta.yml) file that defines its structure:

```yaml
name: example-shell
description: An example shell plugin for testing
type: shell
version: 1.0.0
commands:
  hello:
    description: Say hello
    usage: Say hello to someone
    options:
      - name: name
        short: n
        description: Name to say hello to
        value: "World"
```

## Plugin Types

### Shell Plugins

Shell plugins execute .sh scripts. Arguments and options are passed directly to the script.

Example:

```bash
#!/bin/bash
NAME="World"
while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--name)
      NAME="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done
echo "Hello, $NAME!"
```

### Ansible Plugins

Ansible plugins execute .yml playbook files with any additional arguments passed to ansible-playbook.

### Go Plugins

Go plugins are compiled into the binary and provide the highest performance and integration.

## Usage

### Building the Application

```bash
go build -o dev-tools
```

### Running Commands

List all available commands:

```bash
./dev-tools help
```

Run a shell plugin:

```bash
./dev-tools example-shell hello -n John
```

Run an Ansible plugin:

```bash
./dev-tools example-ansible ping
```

Run an internal Go plugin:

```bash
./dev-tools example-go --message "Hello from Go"
```

## Development

Run directly with go:

```bash
go run main.go example-shell hello
```

Debug with VS Code: Use the provided launch configuration to debug the application.

## Extending Dev-Tools

### Creating a Shell Plugin

1. Create a new directory in `plugins/shell/`
2. Add a `meta.yml` file describing the plugin
3. Create `.sh` scripts for each command
4. Make scripts executable: `chmod +x *.sh`

### Creating an Ansible Plugin

1. Create a new directory in `plugins/ansible/`
2. Add a `meta.yml` file describing the plugin
3. Create `.yml` playbook files for each command

### Creating a Go Plugin

1. Create a new package in `internal/plugin/go/`
2. Implement the PluginApp interface
3. Register the plugin in `internal/plugin/registry.go`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request
