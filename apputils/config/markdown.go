package config

import (
	"fmt"
	"strings"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	flag "github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/apputils/parameter"
)

func escapeAsterisk(text string) string {
	return strings.ReplaceAll(text, "*", "\\*")
}

func groupNameUpper(g *parameter.ParameterGroup, replaceTopicNames map[string]string) string {
	if topicNameReplaced, exists := replaceTopicNames[g.Name]; exists {
		return topicNameReplaced
	}

	return strings.ToUpper(g.Name[:1]) + g.Name[1:]
}

func groupAnchorName(g *parameter.ParameterGroup) string {
	return strings.ToLower(strings.ReplaceAll(g.BaseName, ".", "_"))
}

func groupTableEntries(g *parameter.ParameterGroup, replaceTopicNames map[string]string) [][]string {
	entries := make([][]string, 0)

	for _, entry := range g.Entries {
		switch v := entry.(type) {
		case *parameter.ParameterGroup:
			topicName := v.Name
			if topicNameReplaced, exists := replaceTopicNames[v.Name]; exists {
				topicName = topicNameReplaced
			}

			if v.Default == nil {
				entries = append(entries, []string{
					fmt.Sprintf("[%s](#%s)", v.Name, groupAnchorName(v)),
					fmt.Sprintf("Configuration for %s", topicName),
					"object",
					"",
				})
			} else {
				entries = append(entries, []string{
					fmt.Sprintf("[%s](#%s)", v.Name, groupAnchorName(v)),
					fmt.Sprintf("Configuration for %s", topicName),
					"array",
					"see example below",
				})
			}

		case *parameter.Parameter:
			entries = append(entries, []string{
				v.Name,
				v.Description,
				v.Type,
				v.DefaultStr,
			})
		default:
			panic(parameter.ErrUnknownEntryType)
		}
	}

	return entries
}

func createMarkdownTables(groups []*parameter.ParameterGroup, replaceTopicNames map[string]string) string {
	var result string
	for i, group := range groups {
		groupNumber := i + 1

		if group.Level == 0 {
			result += fmt.Sprintf("## <a id=\"%s\"></a> %d. %s\n\n", groupAnchorName(group), groupNumber, groupNameUpper(group, replaceTopicNames))
		} else {
			result += fmt.Sprintf("### <a id=\"%s\"></a> %s\n\n", groupAnchorName(group), groupNameUpper(group, replaceTopicNames))
		}

		table, err := markdown.NewTableFormatterBuilder().
			WithPrettyPrint().
			Build("Name", "Description", "Type", "Default value").
			Format(groupTableEntries(group, replaceTopicNames))
		if err != nil {
			panic(err)
		}
		result += fmt.Sprintf("%s\n", escapeAsterisk(table))

		if len(group.SubGroups) > 0 {
			result += createMarkdownTables(group.SubGroups, replaceTopicNames)
		}

		if group.Level == 0 {
			result += "Example:\n"
			result += "\n"
			result += "```json\n"
			result += fmt.Sprintf("  %s\n", prettyPrintParameterGroup(group, "  ", "  "))
			result += "```\n"
			result += "\n"
		}
	}

	return result
}

func GetConfigurationMarkdown(config *configuration.Configuration, flagset *flag.FlagSet, ignoreFlags map[string]struct{}, formatTopicNames map[string]string) string {
	return createMarkdownTables(parameter.ParseConfigParameterGroups(config, flagset, ignoreFlags), formatTopicNames)
}
