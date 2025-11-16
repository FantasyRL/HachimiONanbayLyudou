param(
  [Parameter(Mandatory=$true)][string]$Service,
  [Parameter(Mandatory=$true)][string]$Image,
  [Parameter(Mandatory=$true)][string]$ConfigPath,
  [string]$PortMap
)

$ErrorActionPreference = 'Stop'

if (!(Test-Path $ConfigPath)) {
  Write-Error "ERROR: $ConfigPath not found. Please create it."
  exit 2
}

try { docker rm -f $Service *> $null } catch {}

# Default port map by service if not provided
if (-not $PortMap) {
  switch ($Service) {
    'host' { $PortMap = '10001:10001' }
    'mcp_server' { $PortMap = '10002:10002' }
    default { $PortMap = '' }
  }
}

$argsList = @('run','--rm','-itd','--name', $Service)
if ($PortMap) { $argsList += @('-p', $PortMap) }
$argsList += @(
  '-e', ("SERVICE=" + $Service),
  '-e', 'TZ=Asia/Shanghai',
  '-v', ($ConfigPath + ':/app/config/config.yaml:ro'),
  $Image
)

Write-Host ">> docker " ($argsList -join ' ')
& docker @argsList
exit $LASTEXITCODE
