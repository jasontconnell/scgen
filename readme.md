# Sitecore Code Generation and De/Serialization Tool
##### (Name suggestions welcome. The exe is just scgen.exe)

## Configuration

### 'Readonly Configurations'
These should just be in there and not changed.
1. _template_:  The ID for the Template template (string, Guid)
2. _templateFolder_: The ID for the Template Folder template (string, Guid)
3. _templateSection_: The ID for the Template Section template (string, Guid)
4. _templateField_: The ID for the Template Field template (string, Guid)

### Project based configurations
1. _connectionString_: Database connection string. Usually to Sitecore's "master" database. Use "server" instead of "data source". There's no restriction on which database you use. (e.g. user id=sa;password=pwd;server=localhost\\MSSQL_2014;database=Sitecore_master ). 
2. _fieldTypes_: This is a list of how you would like Sitecore field types to be represented in code.
   * typeName: Sitecore type name, e.g. Single Line Text
   * codeType: How this type should be represented in code, e.g. List<Guid>
   * suffix: For some field types, it's useful to add a suffix, for instance I use "ID" for items that qualify to a Guid. Or IDs for List of Guids.
   * ex. { "typeName": "Treelist", "codeType": "List<Guid>", "suffix": "IDs" }
3. _defaultFieldType_: Default type to use if the field type isn't found. Default "string"
4. _basePaths_: This is the list of all base paths that should be included in de/serialization
   * ex. "basePaths": [ "/sitecore/templates/User Defined", "/sitecore/layout/Layouts", "/sitecore/layout/Renderings", "/sitecore/templates/Branches/User Defined" ]

**To enable any feature you just need to provide _true_ for the appropriate setting**

### Serialization configuration
1. _serialize_: Serialize items from the database. All of the serialization settings should be provided. (bool. True or false. Not provided is the same as false)
2. _serializationPath_: Output path for serialization
3. _serializationExtension_: Extension to use for serialized files. Default ".txt"
4. _serializationIgnoredFields_:  Fields to ignore when serializing, deserializing items to/from disk. (Array of string)


### Code Generation configuration
1. _generate_: Generate code for templates in the database. It will search all items in "basePaths" for Templates
2. _filemode_: Generate one file or generate a directory structure based on sitecore item hierarchy of templates. (string, "one" or "many")
3. _outputPath_: If filemode is "one", this should be a file. If "many", it should be a directory. One file is preferred in a C# setting since files would have to be added to the csproj file each time a new template was created.
4. _codeTemplate_: This is the path to the Go text template. You can generally use the same one for each filemode.
5. _codeFileExtension_: This is only used in "many" file mode. No period is needed. (e.g. "cs")
6. _templatePaths_: The template paths and their respective namespaces
   * path, namespace, alternateNamespace, ignore
   * Path and Namespace are pretty straightforward.
   * alternateNamespace is used when generating something other than data model classes or interfaces. Like controllers or view models. You would want the namespace to be generated the same for the data models, but for the view model you would want a different namespace.
   * ignore is used when you need those templates available still, as they are referenced by non-ignored templates still. We don't want to generate code for them, but they're still referenced and should be there.
   * ignore is also helpful when you are generating controllers, not every template has a rendering.
   * It turns out trying to generate more than just the data model and view model is a bit rough. It would need a way to get paths to views. Some things are better left not generated. It would blow up the configuration and the complexity of the application.

### Deserialization configuration
**This will generally use the same configuration as serialization, just work in updating the database instead of serializing to disk**
1. _deserialize_: Turn on deserialization. This will update the database pointed to by connectionString with the paths that the tool finds to need updating. Deletes, Updates and Inserts are all possible. It will only perform these operations on items that need it.

### Remapping configuration
**Remapping is an advanced feature and pretty specific to my needs. We have a tree clone tool which will clone a tree then remap all ids. However we wanted to clone the templates and have the new tree, which was a content tree, be set to the new templates and fields. The remap functionality accomplishes that.**

1. _remap_: Run the remap?  Bool
2. _remapApplyPath_: The path to apply the remap settings.
3. _remapSettings_: The collection of the original path, the cloned path, the original prefix and the cloned prefix.
   * In our sitecore instance, we cloned the templates and renderings, and cloned the original content tree which was using all of the old templates and renderings. The templates and renderings were renamed from something like "Old Site Core" to "New Site Core". Prefix was set to say "remove old site core from the old templates and renderings, and remove New Site Core from the new templates and renderings, then check the names, you should find a match for each template and rendering :)"  Once items are mapped, you just loop through the new site and update template ids and renderings on each item in the new tree. Done.

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