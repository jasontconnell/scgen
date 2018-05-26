package processor

import (
	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/data"
	"github.com/jasontconnell/utility"
	"path"
	"scgen/conf"
	"scgen/model"
	"sort"
	"strings"
)

func updateTemplateNamespaces(cfg conf.Configuration, templates []*model.Template) {
	for _, template := range templates {
		mostMatch := ""
		for _, templatePath := range cfg.TemplatePaths {
			if strings.HasPrefix(template.Path, templatePath.Path) && len(template.Path) > len(mostMatch) {
				mostMatch = templatePath.Path
			}
		}

		rootPath, _ := path.Split(mostMatch)
		rootPath = string(template.Path[len(rootPath)-1:])

		nospaces := strings.Replace(strings.TrimSuffix(rootPath, "/"), " ", "", -1)
		nospaces = strings.Replace(nospaces, "-", "", -1)
		topFolder := path.Dir(nospaces)
		template.Namespace = strings.Replace(topFolder, "/", ".", -1)
		ns := ""
		ans := ""
		staticns := false
		for _, templatePath := range cfg.TemplatePaths {
			if strings.HasPrefix(template.Path, templatePath.Path) {
				ns = templatePath.Namespace
				ans = templatePath.AlternateNamespace
				staticns = templatePath.StaticNamespace
			}
		}
		rns := template.Namespace

		if !staticns {
			template.Namespace = ns + strings.Replace(rns, "/", ".", -1)
		} else {
			template.Namespace = ns
		}
		template.AlternateNamespace = ans + strings.Replace(rns, "/", ".", -1)
	}
}

func updateReferencedNamespaces(cfg conf.Configuration, templates []*model.Template) {
	for _, template := range templates {
		template.ReferencedNamespaces = []string{}
		for _, base := range template.BaseTemplates {
			if base.Namespace != template.Namespace && !utility.HasString(template.ReferencedNamespaces, base.Namespace) {
				template.ReferencedNamespaces = append(template.ReferencedNamespaces, base.Namespace)
			}
		}
	}
}

func getFieldType(cfg conf.Configuration, field data.TemplateFieldNode) (string, string) {
	key := strings.ToLower(field.GetType())
	var fieldType conf.FieldType
	var codeType string
	var ok bool
	if fieldType, ok = cfg.FieldTypeMap[key]; ok {
		codeType = fieldType.CodeType
	} else {
		codeType = cfg.DefaultFieldType
	}
	return codeType, fieldType.Suffix
}

func mapTemplates(cfg conf.Configuration, nodes []data.TemplateNode) map[uuid.UUID]*model.Template {
	m := make(map[uuid.UUID]*model.Template, len(nodes))
	for _, node := range nodes {
		m[node.GetId()] = &model.Template{ID: node.GetId(), Path: node.GetPath(), Name: node.GetName(), CleanName: getCleanName(node.GetName())}
	}

	for _, node := range nodes {
		template := m[node.GetId()]
		for _, bt := range node.GetBaseTemplates() {
			base := m[bt.GetId()]
			template.BaseTemplates = append(template.BaseTemplates, base)
		}

		for _, f := range node.GetFields() {
			codeType, suffix := getFieldType(cfg, f)
			tfield := model.Field{ID: f.GetId(), Name: f.GetName(), CleanName: getCleanName(f.GetName()), FieldType: f.GetType(), CodeType: codeType, Suffix: suffix}
			template.Fields = append(template.Fields, &tfield)
		}
	}

	return m
}

func populateTemplate(cfg conf.Configuration, template *model.Template, includeMap map[uuid.UUID]bool, ignoreMap map[uuid.UUID]bool) {
	ignore := ignoreMap[template.ID]
	include := includeMap[template.ID]

	template.Include = include
	template.Ignore = ignore
	template.Generate = template.Include && !template.Ignore

	for _, b := range template.BaseTemplates {
		populateTemplate(cfg, b, includeMap, ignoreMap)
	}

	for i := len(template.BaseTemplates) - 1; i >= 0; i-- {
		b := template.BaseTemplates[i]
		if !b.Include && b.Ignore {
			template.BaseTemplates = append(template.BaseTemplates[:i], template.BaseTemplates[i+1:]...)
		}
	}
}

func filterTemplates(cfg conf.Configuration, nodes []data.TemplateNode) []*model.Template {
	m := mapTemplates(cfg, nodes)
	ignoreMap := make(map[uuid.UUID]bool)
	includeMap := make(map[uuid.UUID]bool)
	filtered := []*model.Template{}
	for _, template := range m {
		include := false
		ignore := true
		for _, templatePath := range cfg.TemplatePaths {
			if !include && strings.HasPrefix(template.Path, templatePath.Path) {
				include = true
				ignore = templatePath.Ignore
				break
			}
		}

		ignoreMap[template.ID] = ignore
		includeMap[template.ID] = include

		if include {
			filtered = append(filtered, template)
		}
	}

	list := []*model.Template{}

	for _, template := range filtered {
		populateTemplate(cfg, template, includeMap, ignoreMap)

		if template.Include && !template.Ignore {
			sort.Slice(template.Fields, func(i, j int) bool {
				return template.Fields[i].CleanName < template.Fields[j].CleanName
			})
			list = append(list, template)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Path < list[j].Path
	})

	updateTemplateNamespaces(cfg, list)
	updateReferencedNamespaces(cfg, list)

	return list
}
