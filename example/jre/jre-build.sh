#!/usr/bin/env bash

# Define blacklisted modules as a multiline string
BLACKLIST=(
    java.desktop
    java.smartcardio
    jdk.accessibility
    jdk.attach
    jdk.compiler
    jdk.editpad
    jdk.graal.compiler
    jdk.graal.compiler.management
    jdk.hotspot.agent
    jdk.internal.ed
    jdk.internal.jvmstat
    jdk.internal.le
    jdk.internal.opt
    jdk.jartool
    jdk.javadoc
    jdk.jcmd
    jdk.jconsole
    jdk.jdeps
    jdk.jdi
    jdk.jlink
    jdk.jpackage
    jdk.jshell
    jdk.jstatd
    jdk.random
    jdk.rmic
    jdk.unsupported.desktop
)

# Convert the multiline BLACKLIST into a regex pattern with '|'
BLACKLIST_PATTERN=$(IFS="|"; echo "${BLACKLIST[*]}")

# Get all available modules, remove versions, filter out blacklisted ones
MODULES=$(/usr/lib/jvm/default-jvm/bin/java --list-modules | sed 's/@.*//' | grep -v -E "$BLACKLIST_PATTERN")

# Join the module list into a comma-separated string for jlink
MODULES_CSV=$(echo "$MODULES" | tr '\n' ',' | sed 's/,$//')

# Parse the --output parameter if provided
OUTPUT_DIR="jre" # Default output directory
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --output) OUTPUT_DIR="$2"; shift ;;
    esac
    shift
done

# Detect Java version
JAVA_VERSION=$(/usr/lib/jvm/default-jvm/bin/java -version 2>&1 | awk -F '"' '/version/ {print $2}' | cut -d. -f1)

# Decide compression option based on Java version
if [[ "$JAVA_VERSION" -ge 21 ]]; then
    COMPRESS_OPTION="zip-9"
else
    COMPRESS_OPTION="2"
fi

# Use jlink to create the custom runtime image
jlink \
    --add-modules $MODULES_CSV \
    --strip-debug \
    --compress $COMPRESS_OPTION \
    --no-header-files \
    --no-man-pages \
    --output $OUTPUT_DIR

echo "Custom JRE created at: $OUTPUT_DIR"
echo "With modules: $MODULES_CSV"
