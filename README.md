<h1 align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/raito-io/raito-io.github.io/raw/master/assets/images/logo-vertical-dark%402x.png">
    <img height="250px" src="https://github.com/raito-io/raito-io.github.io/raw/master/assets/images/logo-vertical%402x.png">
  </picture>
</h1>

<h4 align="center">
  DBT plugin for the Raito CLI
</h4>

<p align="center">
    <a href="/LICENSE.md" target="_blank"><img src="https://img.shields.io/badge/license-Apache%202-brightgreen.svg" alt="Software License" /></a>
    <a href="https://github.com/raito-io/cli-plugin-dbt/actions/workflows/build.yml" target="_blank"><img src="https://img.shields.io/github/actions/workflow/status/raito-io/cli-plugin-dbt/build.yml?branch=main" alt="Build status"/></a>
    <a href="https://codecov.io/gh/raito-io/cli-plugin-dbt" target="_blank"><img src="https://img.shields.io/codecov/c/github/raito-io/cli-plugin-dbt" alt="Code Coverage" /></a>
</p>

<hr/>

# Raito CLI Plugin - DBT


**Note: This repository is still in an early stage of development.
At this point, no contributions are accepted to the project yet.**


### Prerequisites
To use this plugin, you will need

1. The Raito CLI to be correctly installed. You can check out our [documentation](http://docs.raito.io/docs/cli/installation) for help on this.
2. A Raito Cloud account to synchronize your GCP organization with. If you don't have this yet, visit our webpage at (https://raito.io) and request a trial account.
3. Access to the manifest.json file of a DBT project

### Usage
To use the plugin, add the following snippet to your Raito CLI configuration file (`raito.yml`, by default) under the `targets` section:

```yaml
  - name: dbt1
    connector-name: raito-io/cli-plugin-dbt
    data-source-id: <<GCP datasource ID>>   
    identity-store-id: <<GCP identitystore ID>>
    
    manifest: <<dbt manifest.json file path>>
    

```

You will also need to configure the Raito CLI further to connect to your Raito Cloud account, if that's not set up yet.
A full guide on how to configure the Raito CLI can be found on (http://docs.raito.io/docs/cli/configuration).

### Trying it out

As a first step, you can check if the CLI finds this plugin correctly. In a command-line terminal, execute the following command:
```bash
$> raito info raito-io/cli-plugin-dbt
```

This will download the latest version of the plugin (if you don't have it yet) and output the name and version of the plugin, together with all the plugin-specific parameters to configure it.

When you are ready to try out the synchronization for the first time, execute:
```bash
$> raito run
```
This will take the configuration from the `raito.yml` file (in the current working directory) and start a single synchronization.

Note: if you have multiple targets configured in your configuration file, you can run only this target by adding `--only-targets gcp1` at the end of the command.

## Manifest configuration
### Define a grant
Grants can be defined on models, seeds and snapshots. Within the `raito` object, defined in the [meta](https://docs.getdbt.com/reference/resource-configs/meta){:target=_blank} property, a `grant` array can be defined.
A grant can be defined with the following properties:
* **name** (mandatory): The name of the grant. All grants, defined in the dbt project, with the same name will be combined into one Raito Cloud grant.
* **permissions**: Set of permissions that should be granted within this grant on the current resource.
* **global_permissions**: Set of global permissions (`Read`, `Write`, `Admin`) that should be granted with this grant on the current resource.
* **category**: The category id of the grant. If not provided, the category will be set to the default category.
* **type**: The technical type of the grant. If not provided, the type will be set to the default type.
* **owners**: List of owners of the filter. The owners can be defined by their email addresses.

### Define a mask
Masks can be defined on the columns of models, seeds and snapshots. Within the `raito` object, defined in the [meta](https://docs.getdbt.com/reference/resource-configs/meta){:target=_blank} property, a `mask` can be defined.
A mask can be defined with the following properties:
* **name** (mandatory): A name of the mask. This name should be unique within the dbt project.
* **type**: The mask type that should be used to mask the data. The possible types are defined within the plugin of the corresponding data source. If no type is defined, the default mask of the plugin will be used.
* **owners**: List of owners of the filter. The owners can be defined by their email addresses.

### Define a filter
Filters can be defined on models, seeds and snapshots. Within the `raito` object, defined in the [meta](https://docs.getdbt.com/reference/resource-configs/meta){:target=_blank} property, a `filter` can be defined.
A filter can be defined with the following properties:
* **name** (mandatory): A name of the filter. This name should be unique within the dbt project.
* **policy_rule**: Sql statement defining the filter policy. The policy rule should return a boolean value. If the value is `true`, the data will be included in the result set. If the value is `false`, the data will be excluded from the result set.
* **owners**: List of owners of the filter. The owners can be defined by their email addresses.