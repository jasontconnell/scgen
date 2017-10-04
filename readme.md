# Sitecore Code Generation and De/Serialization Tool
##### (Name suggestions welcome. The exe is just scgen.exe)

## Configuration

### 'Readonly Configurations'
These should just be in there and not changed.
1. _template_:  The ID for the Template template (string, Guid)
2. _templateFolder_: The ID for the Template Folder template (string, Guid)
3. _templateSection_: The ID for the Template Section template (string, Guid)
4. _templateField_: The ID for the Template Field template (string, Guid)

### More configurations that won't change much
1. _serializationIgnoredFields_:  Fields to ignore when serializing, deserializing items to/from disk. (Array of string)
2. _serializationExtension_: Extension for serialized items. Should remain consistent once started.  (string, e.g. "txt")
3. _defaultFieldType_: This will depend on code language which will typically be C#. When a field type doesn't match (string, e.g. "string" [C# code type "System.String"]

### Project based configurations
1. _connectionString_: Database connection string. Usually to Sitecore's "master" database. Use "server" instead of "data source". There's no restriction on which database you use. (e.g. user id=sa;password=pwd;server=localhost\\MSSQL_2014;database=Sitecore_master ). 
2. _fieldTypes_: This is a list of how you would like Sitecore field types to be represented in code.
   * typeName: Sitecore type name, e.g. Single Line Text
   * codeType: How this type should be represented in code, e.g. List<Guid>
   * suffix: For some field types, it's useful to add a suffix, for instance I use "ID" for items that qualify to a Guid. Or IDs for List of Guids.
   * ex. { "typeName": "Treelist", "codeType": "List<Guid>", "suffix": "IDs" }
3. _basePaths_: This is the list of all base paths that should be included in generation de/serialization
   * ex. "basePaths": [ "/sitecore/templates/User Defined", "/sitecore/layout/Layouts", "/sitecore/layout/Renderings", "/sitecore/templates/Branches/User Defined" ]

**To enable any feature you just need to provide _true_ for the appropriate setting**

### Serialization configuration
1. _serialize_: Serialize items from the database. All of the serialization settings should be provided. (bool. True or false. Not provided is the same as false)

### Code Generation configuration
1. _generate_: Generate code for templates in the database. It will search all items in "basePaths" for Templates
2. _baseNamespace_: Starting namespace prefix, e.g. DD.Domain.Models.Glass
3. _filemode_: Generate one file or generate a directory structure based on sitecore item hierarchy of templates. (string, "one" or "many")
4. _outputPath_: If filemode is "one", this should be a file. If "many", it should be a directory. One file is preferred in a C# setting since files would have to be added to the csproj file each time a new template was created.
5. _codeTemplate_: This is the path to the Go text template. You can generally use the same one for each filemode.
6. _codeFileExtension_: This is only used in "many" file mode. No period is needed. (e.g. "cs")

### Deserialization configuration
**This will generally use the same configuration as serialization, just work in updating the database instead of serializing to disk**
1. _deserialize_: Turn on deserialization. This will update the database pointed to by connectionString with the paths that the tool finds to need updating. Deletes, Updates and Inserts are all possible. It will only perform these operations on items that need it.

## Configuration Notes
Configuration files can be broken up by function. This way you can run only the tasks you want when you run the program. The only argument to the program is "-c" for config files. This can be a csv list of configs.

For example, if you have all shared configuration data in "shared.json", all project specific configuration data in "project.json", and a "serialize.json" file with the only property being "serialize": true, you can run scgen like this to ONLY serialize data.

scgen -c shared.json,project.json,serialize.json

Similarly if you have "generate": true in "generate.json", as well as generate specific settings (baseNamespace, filemode, outputPath, etc), you can run scgen as follows

scgen -c shared.json,project.json,generate.json

You can separate configs by function and only run what you want, or you can have a config.json with all of the settings, and modify those flags ("generate", "serialize", "deserialize") on the fly to only do specific things.

However, the multiple file configs are the way to go. Then you can create bat files.

generate.bat:
scgen -c shared.json,project.json,generate.json

serialize.bat:
scgen -c shared.json,project.json,serialize.json

update.bat
--use git to pull first
git pull
scgen -c shared.json,project.json,deserialize.json


(These are just examples and not exact syntax for how you should do a bat file.)


Known issues:
1. This thing works well. If you sync against the wrong database, for instance, it will overwrite whatever is in paths you specify, and insert whatever is on disk. SO BE CAREFUL!
2. After deserializing, you will have to touch the web.config or clear the cache in sitecore manually. It uses no part of the Sitecore API so it's not able to refresh the sitecore cache.
   * One idea for this would just be to have a config setting for path to the running web.config and touch it after deserialization happens.