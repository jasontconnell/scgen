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
            cast(Items.ID as varchar(100)) ID, Name, replace(replace(Name, ' ', ''), '-', '') as NameNoSpaces, cast(TemplateID as varchar(100)) TemplateID, cast(ParentID as varchar(100)) ParentID, cast(MasterID as varchar(100)) as MasterID, Items.Created, Items.Updated, isnull(sf.Value, '') as Type, isnull(Replace(Replace(b.Value, '}',''), '{', ''), '') as BaseTemplates
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
                item := &data.Item{ ID: row["ID"].(string), Name: row["Name"].(string), CleanName: row["NameNoSpaces"].(string), TemplateID: row["TemplateID"].(string), ParentID: row["ParentID"].(string), MasterID: row["MasterID"].(string), Created: row["Created"].(time.Time), Updated: row["Updated"].(time.Time), FieldType: row["Type"].(string), BaseTemplates: row["BaseTemplates"].(string) }
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
                where Language = 'en'
                and Version = 1
            union
            select
                ID, ItemId, FieldId, Value, 1, Language, 'UnversionedFields'
            from UnversionedFields
                where Language = 'en'
        )

        select cast(fv.ValueID as varchar(100)) as ValueID, cast(fv.ItemID as varchar(100)) as ItemID, f.Name as FieldName, cast(fv.FieldID as varchar(100)) as FieldID, fv.Value, fv.Version, fv.Language, fv.Source
                from
                    FieldValues fv
                        join Items f
                            on fv.FieldID = f.ID
                where
                    f.Name not in (%[1]v);
    `

    fieldValues := []*data.FieldValue{}

    sqlstr := fmt.Sprintf(sqlfmt, ignore)
    if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
        defer db.Close()

        if records,err := rs.GetResultSet(db, sqlstr); err == nil {
            for _, row := range records {
                fieldValue := &data.FieldValue{
                        FieldValueID: row["ValueID"].(string),
                        ItemID: row["ItemID"].(string), 
                        FieldName: row["FieldName"].(string),
                        FieldID: row["FieldID"].(string), 
                        Value: row["Value"].(string),  
                        Language: row["Language"].(string), 
                        Version: row["Version"].(int64), 
                        Source: row["Source"].(string) }
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

func deserialize(cfg conf.Configuration, items []data.UpdateItem, fields []data.UpdateField) int64 {
    var updated int64 = 0
    if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
        defer db.Close()

        for _, sql := range getSqlForItems(items) {
            if result, err := db.Exec(sql); err == nil {
                i,_ := result.RowsAffected()
                updated += i
            } else {
                fmt.Println(err)
                return -1
            }
        }

        for _, sql := range getSqlForFields(fields){
            if result, err := db.Exec(sql); err == nil {
                i,_ := result.RowsAffected()
                updated += i
            } else {
                fmt.Println(err)
                return -1
            }
        }
    }

    return updated
}

var updateitemfmt string = "update Items set Name = '%[1]v', TemplateID = '%[2]v', ParentID = '%[3]v', MasterID = '%[4]v' where ID = '%[5]v'"
var insertitemfmt string = "insert into Items (ID, Name, TemplateID, ParentID, MasterID, Created, Updated, DAC_index) values ('%[5]v', '%[1]v', '%[2]v', '%[3]v', '%[5]v', getdate(), getdate(), null)"
var deleteitemfmt string = "delete from Items where ID = '%v'"

func getSqlForItems(items []data.UpdateItem) []string {
    sqllist := []string{}
    for _, item := range items {
        var sql string
        switch item.UpdateType {
        case data.Update:
            sql = fmt.Sprintf(updateitemfmt, item.Name, item.TemplateID, item.ParentID, item.MasterID, item.ID)
        case data.Insert:
            sql = fmt.Sprintf(insertitemfmt, item.Name, item.TemplateID, item.ParentID, item.MasterID, item.ID)
        case data.Delete:
            sql = fmt.Sprintf(deleteitemfmt, item.ID)
        }

        if len(sql) > 0 {
            sqllist = append(sqllist, sql)
        }
    }
    return sqllist
}


func getSqlForFields(fields []data.UpdateField) []string {
    updatemap := make(map[string]string)
    insertmap := make(map[string]string)
    deletemap := make(map[string]string)

    updatemap["SharedFields"] = "update %[1]v set Value = '%[4]v', Updated = getdate() where ItemID = '%[2]v' and FieldID = '%[3]v'"
    updatemap["UnversionedFields"] = "update %[1]v set Value = '%[4]v', Updated = getdate() where ItemID = '%[2]v' and FieldID = '%[3]v' and Language = '%[5]v'"
    updatemap["VersionedFields"] = "update %[1]v set Value = '%[4]v', Updated = getdate() where ItemID = '%[2]v' and FieldID = '%[3]v' and Language = '%[5]v' and Version = %[6]v"

    insertmap["SharedFields"] = "insert into %[1]v (ID, ItemID, FieldID, Value, Created, Updated, DAC_index) values (newid(), '%[2]v', '%[3]v', '%[4]v', getdate(), getdate(), null)"
    insertmap["UnversionedFields"] = "insert into %[1]v (ID, ItemID, FieldID, Value, Language, Created, Updated, DAC_index) values (newid(), '%[2]v', '%[3]v', '%[4]v', '%[5]v', getdate(), getdate(), null)"
    insertmap["VersionedFields"] = "insert into %[1]v (ID, ItemID, FieldID, Value, Language, Version, Created, Updated, DAC_index) values (newid(), '%[2]v', '%[3]v', '%[4]v', '%[5]v', '%[6]v', getdate(), getdate(), null)"

    deletemap["SharedFields"] = "delete from %[1]v where ItemID = '%[2]v' and FieldID = '%[3]v'"
    deletemap["UnversionedFields"] = "delete from %[1]v where ItemID = '%[2]v' and FieldID = '%[3]v' and Language = '%[5]v'"
    deletemap["VersionedFields"] = "delete from %[1]v where ItemID = '%[2]v' and FieldID = '%[3]v' and Language = '%[5]v' and Version = %[6]v"

    sqllist := []string{}
    for _, field := range fields {
        var sql string
        value := strings.Replace(field.Value, "'", "''", -1)

        switch field.UpdateType {
        case data.Update:
            sqlfmt,_ := updatemap[field.Source]
            sql = fmt.Sprintf(sqlfmt, field.Source, field.ItemID, field.FieldID, value, field.Language, field.Version)
        case data.Insert:
            sqlfmt,_ := insertmap[field.Source]
            sql = fmt.Sprintf(sqlfmt, field.Source, field.ItemID, field.FieldID, value, field.Language, field.Version)
        case data.Delete:
            sqlfmt,_ := deletemap[field.Source]
            sql = fmt.Sprintf(sqlfmt, field.Source, field.ItemID, field.FieldID, value, field.Language, field.Version)
        }

        if len(sql) > 0 {
            sqllist = append(sqllist, sql)
        }
    }
    return sqllist
}