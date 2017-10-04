package processor

import (
    "fmt"
    "database/sql"
    _ "github.com/denisenkom/go-mssqldb"
    "scgen/rs"
    "scgen/data"
    "time"
    "scgen/conf"
    "strings"
)

var timefmt string = "2006-01-02 15:04:05.999"

func getItemsForGeneration(cfg conf.Configuration) ([]*data.Item, error) {
    sqlfmt := `
        select 
            cast(Items.ID as varchar(100)) ID, Name, replace(replace(Name, ' ', ''), '-', '') as NameNoSpaces, cast(TemplateID as varchar(100)) TemplateID, cast(ParentID as varchar(100)) ParentID, Items.Created, Items.Updated, isnull(sf.Value, '') as Type, isnull(Replace(Replace(b.Value, '}',''), '{', ''), '') as BaseTemplates
        from
            Items
                left join SharedFields sf
                    on Items.ID = sf.ItemId
                        and sf.FieldId = 'AB162CC0-DC80-4ABF-8871-998EE5D7BA32'
                left join SharedFields b
                    on Items.ID = b.ItemID
                        and b.FieldId = '12C33F3F-86C5-43A5-AEB4-5598CEC45116'
        order by ParentID
    `

    sqlstr := fmt.Sprintf(sqlfmt)
    items := []*data.Item{}

    if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
        defer db.Close()

        if records,err := rs.GetResultSet(db, sqlstr); err == nil {
            for _, row := range records {
                item := &data.Item{ ID: row["ID"].(string), Name: row["Name"].(string), CleanName: row["NameNoSpaces"].(string), TemplateID: row["TemplateID"].(string), ParentID: row["ParentID"].(string), Created: row["Created"].(time.Time), Updated: row["Updated"].(time.Time), FieldType: row["Type"].(string), BaseTemplates: row["BaseTemplates"].(string) }
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

func getItemsForSerialization(cfg conf.Configuration) ([]*data.FieldValue, error){
    ignore := "'" + strings.Join(cfg.SerializationIgnoredFields, "','") + "'"
    sqlfmt := `
        with FieldValues (ItemID, FieldID, Value, Version, Language)
        as
        (
            select
                ItemId, FieldId, Value, 1, 'en'
            from SharedFields
            union
            select
                ItemId, FieldId, Value, Version, Language
            from VersionedFields
                where Language = 'en'
                and Version = 1
            union
            select
                ItemId, FieldId, Value, 1, Language
            from UnversionedFields
                where Language = 'en'
        )

        select cast(i.ID as varchar(100)) as ID, i.Name as ItemName, f.Name as FieldName, cast(i.ParentID as varchar(100)) as ParentID, cast(i.TemplateID as varchar(100)) as TemplateID, i.Created, i.Updated, cast(fv.FieldID as varchar(100)) as FieldID, fv.Value
                from
                    Items i
                        join FieldValues fv
                            join Items f
                                on fv.FieldID = f.ID
                            on i.ID = fv.ItemID
                where
                    f.Name not in (%[1]v);
    `

    fieldValues := []*data.FieldValue{}

    sqlstr := fmt.Sprintf(sqlfmt, ignore)
    if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
        defer db.Close()

        if records,err := rs.GetResultSet(db, sqlstr); err == nil {
            for _, row := range records {
                fieldValue := &data.FieldValue{ ID: row["ID"].(string), ItemName: row["ItemName"].(string), FieldName: row["FieldName"].(string), TemplateID: row["TemplateID"].(string), ParentID: row["ParentID"].(string), FieldID: row["FieldID"].(string), Value: row["Value"].(string), Created: row["Created"].(time.Time), Updated: row["Updated"].(time.Time) }
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