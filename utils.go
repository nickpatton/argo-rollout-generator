package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Take any number of structs (Kubernetes objects), convert them to a YAML string with each object separated by the '---' delimiter
func joinKubernetesObjects(object ...interface{}) string {
	var objects []interface{}
	objects = append(objects, object...)

	var yamlDocs []string

	for _, s := range objects {
		// Convert struct to JSON (respecting all of the omitempty field tags)
		jsonBytes, err := json.Marshal(s)
		if err != nil {
			fmt.Printf("error marshaling struct to JSON: %v", err)
		}

		// Convert JSON to YAML
		var yamlObj map[string]interface{}
		err = yaml.Unmarshal(jsonBytes, &yamlObj)
		if err != nil {
			fmt.Printf("error converting JSON to YAML: %v", err)
		}

		yamlBytes, err := yaml.Marshal(yamlObj)
		if err != nil {
			fmt.Printf("error marshaling YAML: %v", err)
		}

		yamlDocs = append(yamlDocs, string(yamlBytes))
	}

	// Join YAML documents with '---'
	yamlContent := strings.Join(yamlDocs, "---\n")
	return yamlContent

}

// Take string (assumed yaml) and write to file
func writeFile(contents string) error {
	file, err := os.Create("manifest.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}
	return nil
}
