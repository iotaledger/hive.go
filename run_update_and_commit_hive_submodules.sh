#!/bin/bash
BASE_MODULE="github.com/iotaledger/hive.go"

# Check if there are any changes in the Git repository
if [[ -n $(git status -s) ]]; then
  echo "ERROR: There are pending changes in the repository. We can't update the dependencies!"
  exit 1
fi

# Run git fetch
git fetch >/dev/null

# Check if the remote branch is set
current_branch=$(git symbolic-ref --short HEAD)
if ! git show-ref --verify --quiet refs/remotes/origin/$current_branch; then
  echo "ERROR: Remote branch \"origin/$current_branch\" doesn't exist! Create it first by pushing to remote!"
  exit 1
fi

# Check if we or remote is up to date
local_commit=$(git rev-parse "$current_branch")
remote_commit=$(git rev-parse "origin/$current_branch")
if ! [ "$local_commit" = "$remote_commit" ]; then
    echo "ERROR: Current remote branch is not up to date with the local branch!"
  exit 1
fi

# Find all submodules by searching for subdirectories containing go.mod files
SUBMODULES=$(find . -type f -name "go.mod" -exec dirname {} \; | sed -e 's/^\.\///' | sort)

# Declare an associative array to store submodule inter-dependencies
declare -A SUBMODULE_INTER_DEPENDENCIES

# Build the dependency graph
for submodule in $SUBMODULES; do
    dependencies=$(grep -E "^\s$BASE_MODULE" "$submodule/go.mod" | awk '{print $1}')
    SUBMODULE_INTER_DEPENDENCIES["$submodule"]="$dependencies"
done

# Function that updates the inter-dependencies in the submodule
update_submodule() {
    local submodule="$1"
    local dependencies="${SUBMODULE_INTER_DEPENDENCIES["$submodule"]}"

    echo "Updating $submodule..."

    # Enter the submodule folder
    pushd "$submodule" >/dev/null
    
    # Get the current commit hash
    current_commit=$(git rev-parse HEAD)
    current_commit_short=$(git rev-parse --short HEAD)

    for dependency in $dependencies; do
        echo "   go get -u $dependency@$current_commit..."
        go get -u "$dependency@$current_commit" >/dev/null
    done

    # Run go mod tidy
    echo "Running go mod tidy..."
    go mod tidy >/dev/null

    # Commit the changes
    commit_message="Update \"$submodule\" submodule to commit \"$current_commit_short\""
    echo "Commiting the changes with commit message \"$commit_message\"..."
    git add go.mod go.sum
    git commit -m "$commit_message"

    # Push to remote repository, so we can reference the new commits in the other submodules
    git push

    # Add some sleep time, so we can reference the new commits in the other modules
    sleep 2

    # Move back to the parent directory
    popd >/dev/null
}

# Create an empty string to store the ordered submodules
order=""

# Function to recursively resolve dependencies and add them to the order array
resolve_dependencies() {
    local submodule_with_base="$1"
    local submodule="${submodule_with_base#${BASE_MODULE}/}"

    local visited="$2"
    local dependencies="${SUBMODULE_INTER_DEPENDENCIES["$submodule"]}"

    if [[ -z "$visited" ]]; then
        visited="$submodule_with_base "
    else
        visited+="$submodule_with_base "
    fi

    for dependency in $dependencies; do
        if [[ ! "$visited" =~ "$dependency" ]]; then
            resolve_dependencies "$dependency" "$visited"
        fi
    done

    if [[ ! "$order" =~ "$submodule_with_base" ]]; then
        order+="$submodule_with_base "
    fi
}

# Resolve the dependencies between the submodules
for submodule in $SUBMODULES; do
    submodule_with_base=$BASE_MODULE/$submodule
    if [[ ! "$order" =~ "$submodule_with_base" ]]; then
        resolve_dependencies "$submodule_with_base"
    fi
done

# Trim leading and trailing spaces
order="${order%"${order##*[![:space:]]}"}"
order="${order#"${order%%[![:space:]]*}"}"

# Update submodules in the correct order
for submodule_with_base in $order; do
    submodule="${submodule_with_base#${BASE_MODULE}/}"
    update_submodule "$submodule"
done

echo "All submodules updated and committed."
