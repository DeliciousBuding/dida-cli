param(
  [Parameter(Mandatory = $true)]
  [string]$OldBinary,

  [Parameter(Mandatory = $true)]
  [string]$NewBinary,

  [string]$CasesPath = "tests/compat/cases.json",

  [switch]$KeepArtifacts
)

$ErrorActionPreference = "Stop"

function Resolve-File([string]$Path, [string]$Name) {
  $resolved = Resolve-Path -LiteralPath $Path -ErrorAction SilentlyContinue
  if (-not $resolved) {
    throw "$Name not found: $Path"
  }
  return $resolved.ProviderPath
}

function Normalize-Text([string]$Text) {
  if ($null -eq $Text) {
    return ""
  }
  return $Text.Replace("`r`n", "`n").Replace("`r", "`n").TrimEnd("`n")
}

function Normalize-Json([string]$Text, [string]$Label) {
  try {
    $value = $Text | ConvertFrom-Json
    $json = Normalize-Text ($value | ConvertTo-Json -Depth 100 -Compress)
    return Normalize-VolatileJson $json
  } catch {
    throw "$Label did not emit valid JSON: $($_.Exception.Message)"
  }
}

function Normalize-VolatileJson([string]$Text) {
  $normalized = $Text
  $normalized = $normalized -replace ':"[0-9a-f]{24}"', ':"<generated-id>"'
  $normalized = $normalized -replace ':"[0-9a-f]{32}"', ':"<generated-id>"'
  $normalized = $normalized -replace '"createdTime":"[^"]+"', '"createdTime":"<generated-time>"'
  $normalized = $normalized -replace '"modifiedTime":"[^"]+"', '"modifiedTime":"<generated-time>"'
  $normalized = $normalized -replace '"sortOrder":\d+', '"sortOrder":<generated-number>'
  return $normalized
}

function Join-WindowsArguments([object[]]$ArgsList) {
  $quoted = foreach ($arg in $ArgsList) {
    $s = [string]$arg
    if ($s -notmatch '[\s"]' -and $s.Length -gt 0) {
      $s
      continue
    }

    $builder = New-Object System.Text.StringBuilder
    [void]$builder.Append('"')
    $slashCount = 0
    foreach ($ch in $s.ToCharArray()) {
      if ($ch -eq '\') {
        $slashCount++
        continue
      }
      if ($ch -eq '"') {
        [void]$builder.Append('\', ($slashCount * 2 + 1))
        [void]$builder.Append('"')
        $slashCount = 0
        continue
      }
      if ($slashCount -gt 0) {
        [void]$builder.Append('\', $slashCount)
        $slashCount = 0
      }
      [void]$builder.Append($ch)
    }
    if ($slashCount -gt 0) {
      [void]$builder.Append('\', ($slashCount * 2))
    }
    [void]$builder.Append('"')
    $builder.ToString()
  }
  return ($quoted -join " ")
}

function Invoke-DidaCase {
  param(
    [string]$Binary,
    [object[]]$ArgsList,
    [string]$ConfigDir,
    [string]$ArtifactPrefix
  )

  New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
  $stdoutPath = "$ArtifactPrefix.stdout.txt"
  $stderrPath = "$ArtifactPrefix.stderr.txt"

  $psi = New-Object System.Diagnostics.ProcessStartInfo
  $psi.FileName = $Binary
  $psi.Arguments = Join-WindowsArguments $ArgsList
  $psi.UseShellExecute = $false
  $psi.RedirectStandardOutput = $true
  $psi.RedirectStandardError = $true
  $psi.CreateNoWindow = $true
  $psi.EnvironmentVariables["DIDA_CONFIG_DIR"] = $ConfigDir
  $psi.EnvironmentVariables["DIDA365_TOKEN"] = ""
  $psi.EnvironmentVariables["DIDA_ALLOW_TOKEN_ARG"] = ""
  $psi.EnvironmentVariables["DIDA365_OPENAPI_CLIENT_ID"] = ""
  $psi.EnvironmentVariables["DIDA365_OPENAPI_CLIENT_SECRET"] = ""
  $psi.EnvironmentVariables["TICKTICK_CLIENT_ID"] = ""
  $psi.EnvironmentVariables["TICKTICK_CLIENT_SECRET"] = ""

  $process = [System.Diagnostics.Process]::Start($psi)
  $stdout = $process.StandardOutput.ReadToEnd()
  $stderr = $process.StandardError.ReadToEnd()
  $process.WaitForExit()
  $exitCode = $process.ExitCode

  [System.IO.File]::WriteAllText($stdoutPath, $stdout)
  [System.IO.File]::WriteAllText($stderrPath, $stderr)

  return [pscustomobject]@{
    ExitCode = $exitCode
    Stdout = [System.IO.File]::ReadAllText($stdoutPath)
    Stderr = [System.IO.File]::ReadAllText($stderrPath)
    StdoutPath = $stdoutPath
    StderrPath = $stderrPath
  }
}

$old = Resolve-File $OldBinary "Old binary"
$new = Resolve-File $NewBinary "New binary"
$casesFile = Resolve-File $CasesPath "Case manifest"
$manifest = Get-Content -LiteralPath $casesFile -Raw | ConvertFrom-Json

if (-not $manifest.cases -or $manifest.cases.Count -eq 0) {
  throw "No compatibility cases found in $casesFile"
}

$artifactRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("dida-compat-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $artifactRoot | Out-Null

$failed = 0
$passed = 0

try {
  foreach ($case in $manifest.cases) {
    $caseDir = Join-Path $artifactRoot $case.id
    New-Item -ItemType Directory -Force -Path $caseDir | Out-Null

    $argsList = @()
    if ($case.args) {
      $argsList = @($case.args | ForEach-Object { [string]$_ })
    }

    $oldResult = Invoke-DidaCase `
      -Binary $old `
      -ArgsList $argsList `
      -ConfigDir (Join-Path $caseDir "old-config") `
      -ArtifactPrefix (Join-Path $caseDir "old")

    $newResult = Invoke-DidaCase `
      -Binary $new `
      -ArgsList $argsList `
      -ConfigDir (Join-Path $caseDir "new-config") `
      -ArtifactPrefix (Join-Path $caseDir "new")

    $caseFailed = $false
    $expectedExit = [int]$case.expectExit

    if ($oldResult.ExitCode -ne $expectedExit) {
      Write-Host "$($case.id): old binary exit $($oldResult.ExitCode), expected $expectedExit"
      $caseFailed = $true
    }
    if ($newResult.ExitCode -ne $expectedExit) {
      Write-Host "$($case.id): new binary exit $($newResult.ExitCode), expected $expectedExit"
      $caseFailed = $true
    }
    if ($oldResult.ExitCode -ne $newResult.ExitCode) {
      Write-Host "$($case.id): exit mismatch old=$($oldResult.ExitCode) new=$($newResult.ExitCode)"
      $caseFailed = $true
    }

    if ($case.stdoutJson) {
      $oldStdout = Normalize-Json $oldResult.Stdout "$($case.id) old stdout"
      $newStdout = Normalize-Json $newResult.Stdout "$($case.id) new stdout"
    } else {
      $oldStdout = Normalize-Text $oldResult.Stdout
      $newStdout = Normalize-Text $newResult.Stdout
    }

    $oldStderr = Normalize-Text $oldResult.Stderr
    $newStderr = Normalize-Text $newResult.Stderr

    if ($oldStdout -ne $newStdout) {
      Write-Host "$($case.id): stdout mismatch. old=$($oldResult.StdoutPath) new=$($newResult.StdoutPath)"
      $caseFailed = $true
    }
    if ($oldStderr -ne $newStderr) {
      Write-Host "$($case.id): stderr mismatch. old=$($oldResult.StderrPath) new=$($newResult.StderrPath)"
      $caseFailed = $true
    }

    if ($caseFailed) {
      $failed++
      Write-Host "FAIL $($case.id) $($case.category)"
      Write-Host "  args: $($argsList -join ' ')"
      Write-Host "  artifacts: $caseDir"
    } else {
      $passed++
      Write-Host "PASS $($case.id)"
    }
  }

  Write-Host "Compatibility cases passed: $passed"
  if ($failed -gt 0) {
    Write-Host "Compatibility cases failed: $failed"
    Write-Host "Artifacts kept at: $artifactRoot"
    exit 1
  }
} finally {
  if (-not $KeepArtifacts -and $failed -eq 0) {
    Remove-Item -LiteralPath $artifactRoot -Recurse -Force
  } else {
    Write-Host "Artifacts: $artifactRoot"
  }
}
