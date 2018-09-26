go run scripts/cli-structure/generate-cli-structure.go
odo --alsologtostderr --log_backtrace_at --log_dir --logtostderr --skip-connection-check --stderrthreshold --v --vmodule : Odo (Openshift Do)
    app --short : Perform application operations
        create : Create an application
        delete --force : Delete the given application
        describe : Describe the given application
        get --short : Get the active application
        list : List all applications in the current project
        set : Set application as active
    catalog : Catalog related operations
        list : List all available component & service types.
            components : List all components available.
            services : Lists all available services
        search : Search available component & service types.
            component : Search component type in catalog
            service : Search service type in catalog
    component --short : Components of application.
        get --short : Get currently active component
        set : Set active component.
    create --binary --git --local --port : Create a new component
    delete --force : Delete an existing component
    describe : Describe the given component
    link --component : Link target component to source component
    list : List all components in the current application
    log --follow : Retrieve the log for the given component.
    project --short : Perform project operations
        create : Create a new project
        delete --force : Delete a project
        get --short : Get the active project
        list : List all the projects
        set --short : Set the current active project
    push --local : Push source code to a component
    service : Perform service catalog operations
        create : Create a new service
        delete --force : Delete an existing service
        list : List all services in the current application
    storage : Perform storage operations
        create --component --path --size : Create storage and mount to a component
        delete --force : Delete storage from component
        list --all --component : List storage attached to a component
        mount --component --path : mount storage to a component
        unmount --component : Unmount storage from the given path or identified by its name, from the current component
    update --binary --git --local : Update the source code path of a component
    url : Expose component to the outside world
        create --application --component --port : Create a URL for a component
        delete --component --force : Delete a URL
        list --application --component : List URLs
    utils : Utilities for completion, terminal commands and modifying Odo configurations
        completion : Output shell completion code
        config : Modifies configuration settings
            set : Set a value in odo config file
            view : View current configuration values
        terminal : Add Odo terminal support to your development environment
    version : Print the client version information
    watch --delay --ignores : Watch for changes, update component on change
