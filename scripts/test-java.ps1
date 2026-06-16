param(
    [int]$TimeoutSeconds = 120
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$javaDir = Join-Path $root "libs\java"
$sources = Get-ChildItem -Path (Join-Path $javaDir "src\main\java") -Recurse -Filter *.java | ForEach-Object { $_.FullName }
$outDir = Join-Path $env:TEMP "bytemsg233-java-classes"
$smokeFile = Join-Path $env:TEMP "ByteMsg233JavaSmoke.java"

$javaHome = [Environment]::GetEnvironmentVariable("JAVA_HOME", "Process")
if (-not $javaHome) { $javaHome = [Environment]::GetEnvironmentVariable("JAVA_HOME", "Machine") }
if (-not $javaHome) { $javaHome = [Environment]::GetEnvironmentVariable("JAVA_HOME", "User") }
if ($javaHome) {
    $env:JAVA_HOME = $javaHome
    $env:Path = "$javaHome\bin;$env:Path"
}

if (Test-Path $outDir) {
    Remove-Item -Recurse -Force $outDir
}
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

function Invoke-WithTimeout {
    param(
        [string]$FilePath,
        [string[]]$ArgumentList,
        [string]$WorkingDirectory
    )

    $process = Start-Process -FilePath $FilePath -ArgumentList $ArgumentList -WorkingDirectory $WorkingDirectory -NoNewWindow -PassThru -Wait:$false
    if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
        Stop-Process -Id $process.Id -Force
        throw "$FilePath timed out after $TimeoutSeconds seconds."
    }
    if ($process.ExitCode -ne 0) {
        throw "$FilePath exited with code $($process.ExitCode)."
    }
}

$javac = Get-Command javac -ErrorAction SilentlyContinue
if ($javac) {
    & $javac.Source --release 17 -d $outDir @sources
    @"
import com.neko233.bytemsg233.*;
import java.util.*;

public final class ByteMsg233JavaSmoke {
    public static void main(String[] args) {
        ByteMsgWriter writer = new ByteMsgWriter();
        writer.writePackedVarints(Arrays.asList(1L, 2L, 127L, 128L));
        writer.writeDeltaVarints(Arrays.asList(100L, 101L, 109L));
        writer.writeBoolBitset(Arrays.asList(true, false, true, true, false, true, false, false, true));
        writer.writeStringList(Arrays.asList("rank", "battle"));

        ByteMsgReader reader = new ByteMsgReader(writer.toByteArray());
        if (!reader.readPackedVarints(null).equals(Arrays.asList(1L, 2L, 127L, 128L))) throw new RuntimeException("packed failed");
        if (!reader.readDeltaVarints(null).equals(Arrays.asList(100L, 101L, 109L))) throw new RuntimeException("delta failed");
        List<Boolean> flags = reader.readBoolBitset(null);
        if (flags.size() != 9 || !flags.get(0) || flags.get(1) || !flags.get(8)) throw new RuntimeException("bitset failed");
        if (!reader.readStringList(null).equals(Arrays.asList("rank", "battle"))) throw new RuntimeException("strings failed");
    }
}
"@ | Set-Content -Path $smokeFile -Encoding ASCII
    & $javac.Source --release 17 -cp $outDir -d $outDir $smokeFile
    $java = Join-Path (Split-Path -Parent $javac.Source) "java.exe"
    & $java -cp $outDir ByteMsg233JavaSmoke
    Write-Host "Java compile and smoke test passed with local javac."
    exit 0
}

$docker = Get-Command docker -ErrorAction SilentlyContinue
if ($docker) {
    $image = "eclipse-temurin:17-jdk"
    $images = & $docker.Source images --format "{{.Repository}}:{{.Tag}}"
    if ($images -contains $image) {
        Invoke-WithTimeout -FilePath $docker.Source -WorkingDirectory $root -ArgumentList @(
            "run", "--rm",
            "-v", "${root}:/workspace",
            "-w", "/workspace/libs/java",
            $image,
            "bash", "-lc",
            'mkdir -p /tmp/bytemsg233-java && javac --release 17 -d /tmp/bytemsg233-java $(find src/main/java -name "*.java")'
        )
        Write-Host "Java compile passed with Docker $image."
        exit 0
    }
}

throw "No local javac found and Docker image eclipse-temurin:17-jdk is not available. Install JDK 17 or pre-pull the image, then rerun scripts\test-java.ps1."
