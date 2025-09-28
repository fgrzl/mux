Param()

$ErrorActionPreference = 'Stop'

$root = Split-Path -Path $PSScriptRoot -Parent
Set-Location $root

$artifacts = Join-Path $root 'artifacts'
if (-Not (Test-Path $artifacts)) { New-Item -Path $artifacts -ItemType Directory | Out-Null }

function Run-Profile {
    param(
        [string]$Package,
        [string]$BenchName,
        [string]$OutBase
    )

    Write-Host "Profiling package: $Package bench: $BenchName"

    $leaf = Split-Path $Package -Leaf
    $bin = Join-Path $artifacts ("$leaf.test.exe")

    Write-Host "Compiling test binary (for pprof)..."
    & go test -c $Package -o $bin

    $cpuOut = Join-Path $artifacts ("$OutBase.cpu")
    $memOut = Join-Path $artifacts ("$OutBase.mem")
    $runLog = Join-Path $artifacts ("$OutBase.run.txt")

    Write-Host "Running benchmark (go test) and collecting profiles..."
    & go test $Package -bench $BenchName -run '^$' -benchmem -cpuprofile $cpuOut -memprofile $memOut 2>&1 | Tee-Object $runLog

    Write-Host "pprof top (cpu) -> ${OutBase}.cpu.top.txt"
    & go tool pprof -top $bin $cpuOut 2>&1 | Tee-Object (Join-Path $artifacts ("$OutBase.cpu.top.txt"))

    Write-Host "pprof top (mem alloc_objects) -> ${OutBase}.mem.top.txt"
    & go tool pprof -top -alloc_objects $bin $memOut 2>&1 | Tee-Object (Join-Path $artifacts ("$OutBase.mem.top.txt"))

    Write-Host "Artifacts for $OutBase saved to $artifacts"
}

# Heaviest benchmarks to profile
Run-Profile './pkg/router' 'BenchmarkRouter_ManyRoutes_10000' 'router_10000'
Run-Profile './pkg/registry' 'BenchmarkRouteRegistry_Many_10000' 'registry_10000'
Run-Profile './pkg/binder' 'BenchmarkMakeConverter_IntSlice' 'binder_intslice'

Write-Host 'Profiling complete. Inspect artifacts/'
