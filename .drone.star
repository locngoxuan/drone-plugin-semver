def main(ctx):
    version = "1.1.4"

    stages = [
        linux(ctx, "amd64",version),
        linux(ctx, "arm64",version),
        linux(ctx, "arm",version),
    ]

    after = manifest(ctx, version)

    for s in stages:
        for a in after:
            a["depends_on"].append(s["name"])

    return stages + after

def manifest(ctx, version):
    return [{
        "kind": "pipeline",
        "type": "docker",
        "name": "manifest-%s" % (version),
        "steps": [{
            "name":"manifest",
            "image":"plugins/manifest",
            "settings": {
                "target": "xuanloc0511/drone-plugin-semver:%s" % (version),
                "template": "xuanloc0511/drone-plugin-semver:%s-OS-ARCH" % (version),
                "username": {
                    "from_secret": "docker_username",
                },
                "password": {
                    "from_secret": "docker_password",
                },
                "platforms":[
                    "linux/amd64",
                    "linux/arm",
                    "linux/arm64",
                ],
            },            
        }],
        "depends_on": [],
        "trigger": {
            "ref": [
                "refs/heads/main",
                "refs/tags/**",
            ],
        },
    },{
        "kind": "pipeline",
        "type": "docker",
        "name": "manifest-latest",
        "steps": [{
            "name":"manifest",
            "image":"plugins/manifest",
            "settings": {
                "target": "xuanloc0511/drone-plugin-semver:latest",
                "template": "xuanloc0511/drone-plugin-semver:latest-OS-ARCH",
                "username": {
                    "from_secret": "docker_username",
                },
                "password": {
                    "from_secret": "docker_password",
                },
                "platforms":[
                    "linux/amd64",
                    "linux/arm",
                    "linux/arm64",
                ],
            },            
        }],
        "depends_on": [],
        "trigger": {
            "ref": [
                "refs/heads/main",
                "refs/tags/**",
            ],
        },
    }]

def linux(ctx, arch, version):
    build = [
        'go build -v -ldflags "-X main.version=%s" -a -tags netgo -o release/linux/%s/semver .' % (version, arch),
    ]

    steps = [
        {
            "name": "environment",
            "image": "golang:1.16.7",
            "pull": "always",
            "environment": {
                "CGO_ENABLED": "0",
            },
            "commands": [
                "go version",
                "go env",
            ],
        },
        {
            "name": "build",
            "image": "golang:1.16.7",
            "environment": {
                "CGO_ENABLED": "0",
            },
            "commands": build,
        },
        {
            "name": "executable",
            "image": "golang:1.16.7",
            "commands": [
                "./release/linux/%s/semver --help" % (arch),
            ],
        },
    ]

    steps.append({
        "name": "docker",
        "image": "plugins/docker",
        "settings": {
            "dockerfile": "docker/Dockerfile.linux.%s" % (arch),
            "repo": "xuanloc0511/drone-plugin-semver",
            "username": {
                "from_secret": "docker_username",
            },
            "password": {
                "from_secret": "docker_password",
            },
            "tags": [
                "%s-linux-%s" % (version, arch), 
                "latest-linux-%s"% (arch),
            ],
        },
    })

    return {
        "kind": "pipeline",
        "type": "docker",
        "name": "%s-linux-%s" % (version, arch),
        "steps": steps,
        "platform": {
            "os": "linux",
            "arch": arch,
        },
        "depends_on": [],
        "trigger": {
            "ref": [
                "refs/heads/main",
                "refs/tags/**",
                "refs/pull/**",
            ],
        },
    }