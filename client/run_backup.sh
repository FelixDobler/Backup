#!/bin/bash

set -euo pipefail

# load the config file
# configPath="/etc/FDBackup/config.json"
configPath="config.json"
read_config () {
    jq -r "$1" "$configPath"
}

component_property () {
    set +x
    echo $json_properties | jq -r "$@"
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
    # properties=$(read_config ".components.\"$component\" | to_entries[] | \"\(.key)=\"\(.value)\"\n\"")
    json_properties=$(read_config ".components.\"$component\"")
    # echo $properties
    (
        set -x
        # TODO make compatible with properties again
        # export $properties
        export script=$(component_property ".script")
        export json_properties
        export -f component_property

        echo "Running backup script for $component"
        cd $component
        ./$script
        echo "Finished backup script for $component"
        cd ..
    )
done

echo "Finished backup of all components"
