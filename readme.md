# Deployer

## Description description

This is a simple deployer for deploying projects.
The most important part is the synchronisation between local and remote folders.

Actually only `ftp` sync is supported.

Around this sync, there are some other features like:
 - file creation with envs
 - priority file sync
 - http request call
 - file and folder removal
 - file and folder permissions

## Configuration

The application is configured using the config.json file.

The execution is divided into 3 parts:
 - before step
 - sync
 - after step

### Step

Before and after step contain the same steps:
 - generate
 - move
 - action
 - clean

#### Generate

Only one generator is actually supported: `environment_generator`.

##### Config example of generator

You can configure a list of generators. Generator is an object with templatePath and destinationPath.

```json
{
    "type": "environment_generator",
    "arguments":
    {
      "templatePath": "/path_to_template",
      "destination": "/path_to_destination"
    }
}
```

For wildcards, use `{{.ENV_NAME}}` in the template.

#### Move

Move a file from local to remote. Can be a priority uploaded file created in the previous step.

##### Config example of move

Can be configured list of moves. Move is object with source and destination.

```json
{
    "source": "/path_to_source",
    "destination": "/path_to_destination"
}
```

#### Action

Action only supports http/https request calls.

##### Config example of action

Configurable list of actions. Action is an object with url, method, headers and body.

```json
{
  "type": "http_action",
  "arguments":
  {
    "url": "url",
    "method": "http method - GET, POST, PUT, DELETE",
    "headers": {
      "header_name": "header_value"
    },
    "body": "body of request"
  }
}
```

Action also supports wildcards in all arguments. For example, if you want to use the env variable in the url, use `{{.ENV_NAME}}`.

#### Clean

Clean takes two arrays of paths to files and folders to remove. Folder param will remove all files and folders in folder.

##### Config example for clean

Clean is an object with folders and files.

```json
{
  "folders": ["/path_to_folder"],
  "files": ["/path_to_file"]
}
```

### Sync

Sync config contains all information about sync.

**source** - path to the local folder that will be synchronised with the remote folder.

**destination** - path to the remote folder to which the local folder will be synchronised.

**log_file_dest** - path to file where all synced files will be saved - this file increases the sync performance significantly.

**ignore_list** - list of regex that will be ignored during synchronisation. Note that this regex is applied to the whole path, not just the filename.

**default_file_mode** - default file mode for new files. For example 775.

**default_dir_mode** - default dir mode for new dirs. For example 0775.

**ftp_config** - ftp configuration for connection to remote server. Includes host, user and password.

#### Config example for sync

```json
{
  "source": "/path_to_source",
  "destination": "/path_to_destination",
  "log_file_dest": "/path_to_log_file",
  "ignore_list": ["regex1", "regex2"],
  "default_file_mode": "775",
  "default_dir_mode": "0775",
  "ftp_config":
  {
    "host": "host:port",
    "user": "user",
    "password": "{{.FTP_PASSWORD}}"
  }
}
```

### Other configuration

There are some other config parameters that are not necessary, but can be useful.

#### Folders

Folders is an array of paths to folders that will be created after sync.

#### Readable folders

Readable folders is an array of paths to folders that will be set to permission 0777 after sync.

### Whole config example

```json
{
  "before": {
    "generate": [
      {
        "type": "environment_generator",
        "arguments": {
          "templatePath": "/path_to_template",
          "destination": "/path_to_destination"
        }
      }
    ],
    "move": [
      {
        "source": "/path_to_source",
        "destination": "/path_to_destination"
      }
    ],
    "action": [
      {
        "type": "http_action",
        "arguments": {
          "url": "url?param={{.ENV_NAME}}",
          "method": "POST",
          "headers": {
            "header_example": "{{.ENV_NAME}}"
          },
          "body": "{\"example\": \"{{.ENV_NAME}}\"}"
        }
      }
    ],
    "clean": {
      "folders": ["/temp/cache", "/temp/sessions"],
      "files": ["/www/.maintenance.php"]
    }
  },
  "sync": {
    "source": "/source_path",
    "destination": "/destination_path",
    "log_file_dest": "/path_to_log_file",
    "ignore_list": ["regex1", "regex2"],
    "default_file_mode": "775",
    "default_dir_mode": "0775",
    "ftp_config": {
      "host": "host:port",
      "user": "user",
      "password": "{{.FTP_PASSWORD}}"
    }
  },
  "folders": ["example_folder", "example_folder2"],
  "readable_folders": ["/example_folder", "/example_folder2"],
  "after": {
  }
}
```

## Example call

```shell
-- short version
./deployer deploy -c path_to_config -t ftp

-- long version
./deployer deploy --config path_to_config --type ftp
```

## Improvements
- [ ] Use context for cancel call, config will contain timeout
- [ ] Add support for other syncs like SFTP
- [ ] Add support for ordering of steps actions and reusing of steps 