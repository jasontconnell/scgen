package data

import (
    "fmt"
)

func (item SerializedItem) String() string {
    fields := ""
    for _, f := range item.Fields {
        fields += fmt.Sprintf("\nFields\n    Name: %v  Value length: %v", f.Name, len(f.Value))
    }
    s := fmt.Sprintf("Name: %v  ID: %v ", item.Item.Name, item.Item.ID)

    return s + fields
}