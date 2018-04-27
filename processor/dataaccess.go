package processor

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"scgen/conf"
	"scgen/data"
	"strings"
	"time"
    "github.com/jasontconnell/sqlhelp"
)

func getItemsForGeneration(cfg conf.Configuration) ([]*data.Item, error) {
	sqlfmt := `
        select 
            cast(Items.ID as varchar(100)) ID, Name, cast(TemplateID as varchar(100)) TemplateID, cast(ParentID as varchar(100)) ParentID, cast(MasterID as varchar(100)) as MasterID, Items.Created, Items.Updated, isnull(sf.Value, '') as Type, isnull(Replace(Replace(UPPER(b.Value), '}',''), '{', ''), '') as BaseTemplates
        from
            Items
                left join SharedFields sf
                    on Items.ID = sf.ItemId
                        and sf.FieldId = 'AB162CC0-DC80-4ABF-8871-998EE5D7BA32'
                left join SharedFields b
                    on Items.ID = b.ItemID
                        and b.FieldId = '12C33F3F-86C5-43A5-AEB4-5598CEC45116'
            order by Items.Name
    `

	sqlstr := fmt.Sprintf(sqlfmt)
	items := []*data.Item{}

	if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
		defer db.Close()

		if records, err := sqlhelp.GetResultSet(db, sqlstr); err == nil {
			for _, row := range records {
				name := row["Name"].(string)
				cleanName := strings.Replace(strings.Replace(strings.Title(name), "-", "", -1), " ", "", -1)
				if strings.IndexAny(cleanName, "0123456789") == 0 {
					cleanName = "_" + cleanName
				}
				item := &data.Item{ID: row["ID"].(string), Name: name, CleanName: cleanName, TemplateID: row["TemplateID"].(string), ParentID: row["ParentID"].(string), MasterID: row["MasterID"].(string), Created: row["Created"].(time.Time), Updated: row["Updated"].(time.Time), FieldType: row["Type"].(string), BaseTemplates: row["BaseTemplates"].(string)}
				items = append(items, item)
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
	return items, nil
}

func getItemsForSerialization(cfg conf.Configuration) ([]*data.FieldValue, error) {
	ignore := "'" + strings.Join(cfg.SerializationIgnoredFields, "','") + "'"
	sqlfmt := `
        with FieldValues (ValueID, ItemID, FieldID, Value, Version, Language, Source)
        as
        (
            select
                ID, ItemId, FieldId, Value, 1, 'en', 'SharedFields'
            from SharedFields
            union
            select
                ID, ItemId, FieldId, Value, Version, Language, 'VersionedFields'
            from VersionedFields
            union
            select
                ID, ItemId, FieldId, Value, 1, Language, 'UnversionedFields'
            from UnversionedFields
        )

        select cast(fv.ValueID as varchar(100)) as ValueID, cast(fv.ItemID as varchar(100)) as ItemID, f.Name as FieldName, cast(fv.FieldID as varchar(100)) as FieldID, fv.Value, fv.Version, fv.Language, fv.Source
                from
                    FieldValues fv
                        join Items f
                            on fv.FieldID = f.ID
                where
                    f.Name not in (%[1]v)
            order by fv.Source, f.Name, fv.Language, fv.Version;
    `

	fieldValues := []*data.FieldValue{}

	sqlstr := fmt.Sprintf(sqlfmt, ignore)
	if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
		defer db.Close()

		if records, err := sqlhelp.GetResultSet(db, sqlstr); err == nil {
			for _, row := range records {
				fieldValue := &data.FieldValue{
					FieldValueID: row["ValueID"].(string),
					ItemID:       row["ItemID"].(string),
					FieldName:    row["FieldName"].(string),
					FieldID:      row["FieldID"].(string),
					Value:        row["Value"].(string),
					Language:     row["Language"].(string),
					Version:      row["Version"].(int64),
					Source:       row["Source"].(string)}
				fieldValues = append(fieldValues, fieldValue)
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
	return fieldValues, nil
}
