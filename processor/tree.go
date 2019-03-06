package processor

import (
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jasontconnell/scgen/conf"
	"github.com/jasontconnell/scgen/model"
	"github.com/jasontconnell/sitecore/data"
	"github.com/jasontconnell/utility"
)

func updateTemplateNamespaces(cfg conf.Configuration, templates []*model.Template) {
	nameFunc := getCleanNameFunc(cfg.NameStyle)

	for _, template := range templates {
		mostMatch := ""
		for _, templatePath := range cfg.TemplatePaths {
			if strings.HasPrefix(template.Path, templatePath.Path) && len(template.Path) > len(mostMatch) {
				mostMatch = templatePath.Path
			}
		}

		rootPath, _ := path.Split(mostMatch)
		rootPath = string(template.Path[len(rootPath)-1:])

		parts := strings.Split(rootPath, "/")
		for i := range parts {
			parts[i] = nameFunc(parts[i])
		}

		nospaces := strings.TrimSuffix(strings.Join(parts, "/"), "/")
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

func getFieldProperties(cfg conf.Configuration, field data.TemplateFieldNode) map[string]string {
	key := strings.ToLower(field.GetType())
	var fieldType conf.FieldType
	properties := make(map[string]string)
	var ok bool
	if fieldType, ok = cfg.FieldTypeMap[key]; ok && fieldType.Properties != nil {
		properties = fieldType.Properties
	}
	return properties
}

func mapTemplates(cfg conf.Configuration, nodes []data.TemplateNode) map[uuid.UUID]*model.Template {
	nameFunc := getCleanNameFunc(cfg.NameStyle)
	altNameFunc := getCleanNameFunc(cfg.AltNameStyle)

	m := make(map[uuid.UUID]*model.Template, len(nodes))
	for _, node := range nodes {
		m[node.GetId()] = &model.Template{ID: node.GetId(), Path: node.GetPath(), Name: node.GetName(), CleanName: nameFunc(node.GetName()), AllBaseTemplatesMap: make(map[uuid.UUID]*model.Template), AllBaseTemplateIDsMap: make(map[uuid.UUID]bool)}
	}

	for _, node := range nodes {
		template := m[node.GetId()]
		for _, bt := range node.GetBaseTemplates() {
			base := m[bt.GetId()]
			template.BaseTemplates = append(template.BaseTemplates, base)
		}

		for _, f := range node.GetFields() {
			codeType, suffix := getFieldType(cfg, f)
			props := getFieldProperties(cfg, f)
			tfield := model.Field{ID: f.GetId(), Name: f.GetName(), CleanName: nameFunc(f.GetName()), AltCleanName: altNameFunc(f.GetName()), FieldType: f.GetType(), CodeType: codeType, Suffix: suffix, Properties: props}
			template.Fields = append(template.Fields, &tfield)
		}
	}

	for _, node := range m {
		v := make(map[uuid.UUID]bool)
		allBase := getAllBaseTemplateUids(node, v)
		for _, uid := range allBase {
			node.AllBaseTemplateIDsMap[uid] = true
		}
	}

	return m
}

func getAllBaseTemplateUids(node *model.Template, visited map[uuid.UUID]bool) []uuid.UUID {
	list := []uuid.UUID{}
	for _, n := range node.BaseTemplates {
		if _, ok := visited[n.ID]; ok {
			continue
		}
		visited[n.ID] = true

		list = append(list, n.ID)
		b := getAllBaseTemplateUids(n, visited)
		list = append(list, b...)
	}
	return list
}

func mapAll(nodes map[uuid.UUID]*model.Template) {
	for _, node := range nodes {
		all := []*model.Template{}
		for uid := range node.AllBaseTemplateIDsMap {
			t, ok := nodes[uid]
			if ok {
				all = append(all, t)
			}
		}

		sort.Slice(all, func(i, j int) bool {
			return all[i].Path < all[j].Path
		})

		node.AllBaseTemplates = all

		for _, n := range node.AllBaseTemplates {
			n.AllBaseTemplatesMap[n.ID] = n
		}

		fm := make(map[string]bool)
		node.AllFields = append(node.AllFields, node.Fields...)
		for _, f := range node.Fields {
			fm[f.Name] = true
		}

		for _, b := range node.AllBaseTemplates {
			for _, f := range b.Fields {
				if _, ok := fm[f.Name]; ok {
					continue
				}

				node.AllFields = append(node.AllFields, f)
				fm[f.Name] = true
			}
		}
	}
}

func getAllBaseTemplates(node *model.Template, nodes map[uuid.UUID]*model.Template) []*model.Template {
	all := []*model.Template{}

	for id := range node.AllBaseTemplateIDsMap {
		if n, ok := nodes[id]; ok {
			all = append(all, n)
		}
	}
	return all
}

func determineFlags(cfg conf.Configuration, nodes map[uuid.UUID]*model.Template) {
	for _, node := range nodes {
		if cfg.Flags.RenderingParameters {
			_, ok := node.AllBaseTemplateIDsMap[data.RenderingParametersID]
			node.Flags.RenderingParameters = ok
		}
	}
}

func populateTemplate(cfg conf.Configuration, template *model.Template, includeMap map[uuid.UUID]bool, ignoreMap, visited map[uuid.UUID]bool) {
	visited[template.ID] = true
	ignore := ignoreMap[template.ID]
	include := includeMap[template.ID]

	template.Include = include
	template.Ignore = ignore
	template.Generate = template.Include && !template.Ignore

	for _, b := range template.BaseTemplates {
		if _, ok := visited[b.ID]; !ok {
			populateTemplate(cfg, b, includeMap, ignoreMap, visited)
		}
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
	visited := make(map[uuid.UUID]bool)

	for _, template := range filtered {
		populateTemplate(cfg, template, includeMap, ignoreMap, visited)

		if template.Include && !template.Ignore {
			sort.Slice(template.Fields, func(i, j int) bool {
				return template.Fields[i].CleanName < template.Fields[j].CleanName
			})

			sort.Slice(template.AllFields, func(i, j int) bool {
				return template.AllFields[i].CleanName < template.AllFields[j].CleanName
			})
			list = append(list, template)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Path < list[j].Path
	})

	updateTemplateNamespaces(cfg, list)
	updateReferencedNamespaces(cfg, list)

	fmap := make(map[uuid.UUID]*model.Template)
	for _, t := range list {
		if t.Include {
			fmap[t.ID] = t
		}
	}

	if cfg.PopulateAllBaseTemplates {
		mapAll(fmap)
		if cfg.DetermineFlags {
			determineFlags(cfg, fmap)
		}
	}

	return list
}
