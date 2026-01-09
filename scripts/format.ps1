param(
    [switch]$Commit
)

Write-Host "Running gofmt -w on repository..."
& gofmt -w .

if (Get-Command goimports -ErrorAction SilentlyContinue) {
    Write-Host "Running goimports -w (found in PATH)..."
    & goimports -w .
}

$files = & gofmt -l .
if ($files) {
    Write-Host "Files still needing formatting:`n$files"
} else {
    Write-Host "All files formatted."
}

if ($Commit) {
    Write-Host "Committing formatting changes..."
    & git add -A
    & git commit -m "ci: format with gofmt/goimports"
}
