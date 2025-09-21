#!/bin/bash

set -euo pipefail

# load the config file
# configPath="/etc/FDBackup/config.json"
configPath="config.json"
read_config () {
    jq -r "$1" "$configPath"
}

export alert_email=$(read_config ".email")
export rsyncTargetHost=$(read_config ".rsyncTargetHost")

# load enabled components
export enabled_components=$(read_config ".components | to_entries[] | select(.value.enabled) | .key")

# TODO error handling of script and subcomponents
    # TODO email alerts
# TODO Logging of output and errors of subcomponents


export component # make component available to sub-scripts
for component in $enabled_components; do
    echo "--- Component: $component ---"
    echo "Loading configuration for $component"
    properties=$(read_config ".components.\"$component\" | to_entries[] | \"\(.key)=\(.value)\n\"")
    echo $properties
    (
        set -x
        export $properties
        echo "Running backup script for $component"
        $component/$script
        echo "Finished backup script for $component"
    )
done

echo "Finished backup of all components"
