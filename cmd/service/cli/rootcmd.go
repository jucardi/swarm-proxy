package cli

import (
	"fmt"
	"github.com/jucardi/swarm-proxy/cmd/service/version"
	"github.com/jucardi/swarm-proxy/docker"
	"github.com/spf13/cobra"
	"encoding/json"
	"github.com/jucardi/swarm-proxy/proxy"
)

const (
	usage = `%s [template file] -i [JSON or YAML file] -u [URL to GET JSON or INPUT from] -s [JSON or YAML string] -o [output] -p [pattern] -d [template path 1] -d [template path 2]

  - All flags are optional
  - Max one input type allowed (-f, -s, -u)`
	long = `
Infuse - the templates CLI parser
    Version: V-%s
    Built: %s

Supports:
    - Go templates
    - Handlebars templates (coming soon)
`
)

var rootCmd = &cobra.Command{
	Use:              "swarm-proxy",
	Short:            "Starts a proxy service powered by NGINX which serves as an ingress to map services into a single port.",
	Long:             fmt.Sprintf(long, version.Version, version.Built),
	PersistentPreRun: initCmd,
	Run:              start,
}

// Execute starts the execution of the parse command.
func Execute() {
	rootCmd.Flags().StringP("template", "t", "", "the source template")
	rootCmd.Flags().StringP("output", "o", "", "the output of the parsed template")

	rootCmd.Execute()
}

func printUsage(cmd *cobra.Command) {
	cmd.Println(fmt.Sprintf(long, version.Version, version.Built))
	cmd.Usage()
}

func initCmd(cmd *cobra.Command, args []string) {
	FromCommand(cmd)
	cmd.Use = fmt.Sprintf(usage, cmd.Use)
}

func start(cmd *cobra.Command, args []string) {
	//filename := args[0]
	//input, _ := cmd.Flags().GetString("file")
	//str, _ := cmd.Flags().GetString("string")
	//url, _ := cmd.Flags().GetString("url")
	//output, _ := cmd.Flags().GetString("output")
	//definitions, _ := cmd.Flags().GetStringArray("definition")
	//pattern, _ := cmd.Flags().GetString("pattern")

	Test()
}

func Test() {
	containers, _ := docker.Client().GetContainers()
	println("=== CONTAINERS ===")
	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
	services, _ := docker.Client().GetServices()
	println("=== SERVICES ===")
	for _, service := range services {
		labels, _ := json.MarshalIndent(service.Spec.Labels, "  ", "  ")
		fmt.Printf("%s %s \n%s\n", service.ID[:10], service.Spec.Name, labels)
	}
	nodes, _ := docker.Client().GetNodes()
	println("=== NODES ===")
	for _, node := range nodes {
		//labels, _ := json.Marshal(node.Spec.Labels)
		fmt.Printf("%s %s %s\n", node.ID[:10], node.Description.Hostname, node.Spec.Role)
	}

	result, _ := proxy.Service().GetProxyConfig()
	info, _ := json.MarshalIndent(result, "  ", "  ")
	println("=== PROXY INFO ===")
	println(string(info))

	println("=== TEMPLATE ===")
	template, err := proxy.Service().ParseTemplate(result)
	println(template)
	if err != nil {
		println(err.Error())
	}
}
