package processor

import (
    "fmt"
    "database/sql"
    _ "github.com/denisenkom/go-mssqldb"
    "scgen/rs"
    "scgen/data"
    "time"
    "scgen/conf"
)

var timefmt string = "2006-01-02 15:04:05.999"

func getItems(cfg conf.Configuration) ([]*data.Item, error) {
    sqlfmt := `
        select 
            cast(Items.ID as varchar(100)) ID, Name, cast(TemplateID as varchar(100)) TemplateID, cast(ParentID as varchar(100)) ParentID, Items.Created, Items.Updated, isnull(sf.Value, '') as Type, isnull(Replace(Replace(b.Value, '}',''), '{', ''), '') as BaseTemplates
        from
            Items
                left join SharedFields sf
                    on Items.ID = sf.ItemId
                        and sf.FieldId = 'AB162CC0-DC80-4ABF-8871-998EE5D7BA32'
                left join SharedFields b
                    on Items.ID = b.ItemID
                        and b.FieldId = '12C33F3F-86C5-43A5-AEB4-5598CEC45116'
        where
            TemplateID in ('%[1]v','%[2]v','%[3]v','%[4]v')
            Or Items.ID in (select ParentID from Items where TemplateID in ('%[1]v','%[2]v','%[3]v','%[4]v'))
        order by ParentID
    `

    sqlstr := fmt.Sprintf(sqlfmt, cfg.TemplateID, cfg.TemplateFolderID, cfg.TemplateFieldID, cfg.TemplateSectionID)
    items := []*data.Item{}

    if db, err := sql.Open("mssql", cfg.ConnectionString); err == nil {
        defer db.Close()

        if records,err := rs.GetResultSet(db, sqlstr); err == nil {
            for _, row := range records {
                item := &data.Item{ ID: row["ID"].(string), Name: row["Name"].(string), TemplateID: row["TemplateID"].(string), ParentID: row["ParentID"].(string), Created: row["Created"].(time.Time), Updated: row["Updated"].(time.Time), FieldType: row["Type"].(string), BaseTemplates: row["BaseTemplates"].(string) }
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