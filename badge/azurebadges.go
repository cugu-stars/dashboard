package badge

var isAzure = func(p Project) bool { return p.AzureDefinitionID != "" }

func InitAzureBadges() {
	badges["azure-pipeline"] = markdownBadge("https://img.shields.io/azure-devops/build/{{.AzureOrganization}}/{{.AzureProject}}/{{.AzureDefinitionID}}", "https://dev.azure.com/{{.AzureOrganization}}/{{.AzureProject}}/_build?definitionId={{.AzureDefinitionID}}&_a=summary", isAzure)
	badges["azure-coverage"] = markdownBadge("https://img.shields.io/azure-devops/coverage/{{.AzureOrganization}}/{{.AzureProject}}/{{.AzureDefinitionID}}", "{{.URL}}", isAzure)
}
