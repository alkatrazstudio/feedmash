Push-Location "$PSScriptRoot"
try {
    $version = [IO.File]::ReadAllText("VERSION").trim()
    $ts = Get-Date -UFormat %s
    $git_hash = git rev-parse HEAD

    go build -trimpath -a -ldflags "
        -s
        -w
        -X 'feedmash/src.appVersion=$version'
        -X 'feedmash/src.appBuildTimestamp=$ts'
        -X 'feedmash/src.appBuildGitHash=$git_hash'
    "
} finally {
    Pop-Location
}
