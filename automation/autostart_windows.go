//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	if len(os.Args) > 1 { return }
	exe, err := os.Executable()
	if err != nil { showStartupError(err); os.Exit(1) }
	exeDir := filepath.Dir(exe)
	_ = os.Chdir(exeDir)
	const configPath = "automation.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if saveErr := saveConfig(configPath, defaultConfig()); saveErr != nil { showStartupError(fmt.Errorf("create starter config: %w", saveErr)); os.Exit(1) }
	}
	if err := launchNativeControlPanel(exe, filepath.Join(exeDir, configPath)); err != nil { showStartupError(err); os.Exit(1) }
	os.Exit(0)
}

func launchNativeControlPanel(exePath, configPath string) error {
	const script = `
Add-Type -AssemblyName PresentationFramework
Add-Type -AssemblyName PresentationCore
Add-Type -AssemblyName WindowsBase
Add-Type -AssemblyName System.Windows.Forms

$exe = $env:IDLEON_AGENT_EXE
$config = $env:IDLEON_AGENT_CONFIG
$nl = [Environment]::NewLine
$appDir = Split-Path -Parent $exe
$sourceConfig = Join-Path $appDir 'agent-source.json'
$clipboardImport = Join-Path $appDir 'clipboard-account-import.json'

if ([string]::IsNullOrWhiteSpace($exe) -or -not (Test-Path -LiteralPath $exe)) {
    [System.Windows.MessageBox]::Show('Backend executable path is missing or invalid: ' + $exe, 'IdleOn Account Agent') | Out-Null
    exit 2
}

function Quote-Arg([string]$s) {
    if ($null -eq $s) { return '""' }
    return '"' + ($s -replace '"','\"') + '"'
}
function Join-Args([string[]]$items) { return (($items | ForEach-Object { Quote-Arg $_ }) -join ' ') }

$window = New-Object System.Windows.Window
$window.Title = 'IdleOn Account Agent - Analysis Test Build'
$window.Width = 1120
$window.Height = 820
$window.MinWidth = 900
$window.MinHeight = 660
$window.WindowStartupLocation = 'CenterScreen'
$window.Background = '#101214'

$root = New-Object System.Windows.Controls.Grid
$root.Margin = '18'
$root.RowDefinitions.Add((New-Object System.Windows.Controls.RowDefinition -Property @{Height='Auto'}))
$root.RowDefinitions.Add((New-Object System.Windows.Controls.RowDefinition -Property @{Height='Auto'}))
$root.RowDefinitions.Add((New-Object System.Windows.Controls.RowDefinition -Property @{Height='*'}))
$window.Content = $root

$header = New-Object System.Windows.Controls.StackPanel
[System.Windows.Controls.Grid]::SetRow($header,0)
$root.Children.Add($header) | Out-Null

$title = New-Object System.Windows.Controls.TextBlock
$title.Text = 'IdleOn Account Agent'
$title.FontSize = 27
$title.FontWeight = 'Bold'
$title.Foreground = '#F2F2F2'
$header.Children.Add($title) | Out-Null

$subtitle = New-Object System.Windows.Controls.TextBlock
$subtitle.Text = 'Account reading + progression analysis validation build. Gameplay automation is intentionally disabled until parsing is verified.'
$subtitle.TextWrapping = 'Wrap'
$subtitle.Margin = '0,4,0,10'
$subtitle.Foreground = '#AAB4BE'
$header.Children.Add($subtitle) | Out-Null

$sourceBox = New-Object System.Windows.Controls.Border
$sourceBox.BorderBrush = '#30363D'
$sourceBox.BorderThickness = '1'
$sourceBox.Padding = '12'
$sourceBox.Margin = '0,0,0,10'
$sourceBox.Background = '#171B1F'
$header.Children.Add($sourceBox) | Out-Null
$sourceStack = New-Object System.Windows.Controls.StackPanel
$sourceBox.Child = $sourceStack

$sourceTitle = New-Object System.Windows.Controls.TextBlock
$sourceTitle.Text = 'Account Data Source'
$sourceTitle.FontWeight = 'Bold'
$sourceTitle.Foreground = '#F2F2F2'
$sourceTitle.Margin = '0,0,0,4'
$sourceStack.Children.Add($sourceTitle) | Out-Null

$sourceHelp = New-Object System.Windows.Controls.TextBlock
$sourceHelp.Text = 'Use a public IdleOn Toolbox profile, an IdleOn Efficiency Raw Data JSON export, or paste JSON copied from either community tool. Local companion snapshots are used automatically when no source is selected.'
$sourceHelp.TextWrapping = 'Wrap'
$sourceHelp.Foreground = '#8E99A4'
$sourceHelp.Margin = '0,0,0,9'
$sourceStack.Children.Add($sourceHelp) | Out-Null

$toolRow = New-Object System.Windows.Controls.DockPanel
$toolRow.Margin = '0,0,0,7'
$sourceStack.Children.Add($toolRow) | Out-Null
$toolLabel = New-Object System.Windows.Controls.TextBlock
$toolLabel.Text = 'Toolbox public profile / main character:'
$toolLabel.Width = 250
$toolLabel.VerticalAlignment = 'Center'
$toolLabel.Foreground = '#AAB4BE'
[System.Windows.Controls.DockPanel]::SetDock($toolLabel,'Left')
$toolRow.Children.Add($toolLabel) | Out-Null
$toolboxInput = New-Object System.Windows.Controls.TextBox
$toolboxInput.MinWidth = 430
$toolboxInput.Background = '#22272D'
$toolboxInput.Foreground = '#F2F2F2'
$toolboxInput.BorderBrush = '#3A424A'
$toolboxInput.Padding = '6,4'
$toolRow.Children.Add($toolboxInput) | Out-Null

$jsonRow = New-Object System.Windows.Controls.DockPanel
$sourceStack.Children.Add($jsonRow) | Out-Null
$jsonLabel = New-Object System.Windows.Controls.TextBlock
$jsonLabel.Text = 'Raw/community JSON import:'
$jsonLabel.Width = 250
$jsonLabel.VerticalAlignment = 'Center'
$jsonLabel.Foreground = '#AAB4BE'
[System.Windows.Controls.DockPanel]::SetDock($jsonLabel,'Left')
$jsonRow.Children.Add($jsonLabel) | Out-Null

$paste = New-Object System.Windows.Controls.Button
$paste.Content = 'Paste JSON'
$paste.Width = 100
$paste.Margin = '8,0,0,0'
$paste.Background = '#252B31'
$paste.Foreground = '#F2F2F2'
[System.Windows.Controls.DockPanel]::SetDock($paste,'Right')
$jsonRow.Children.Add($paste) | Out-Null

$browse = New-Object System.Windows.Controls.Button
$browse.Content = 'Browse...'
$browse.Width = 90
$browse.Margin = '8,0,0,0'
$browse.Background = '#252B31'
$browse.Foreground = '#F2F2F2'
[System.Windows.Controls.DockPanel]::SetDock($browse,'Right')
$jsonRow.Children.Add($browse) | Out-Null

$jsonInput = New-Object System.Windows.Controls.TextBox
$jsonInput.MinWidth = 390
$jsonInput.Background = '#22272D'
$jsonInput.Foreground = '#F2F2F2'
$jsonInput.BorderBrush = '#3A424A'
$jsonInput.Padding = '6,4'
$jsonRow.Children.Add($jsonInput) | Out-Null

try {
    if (Test-Path -LiteralPath $sourceConfig) {
        $saved = Get-Content -LiteralPath $sourceConfig -Raw | ConvertFrom-Json
        if ($saved.toolbox) { $toolboxInput.Text = [string]$saved.toolbox }
        if ($saved.jsonImport) { $jsonInput.Text = [string]$saved.jsonImport }
        elseif ($saved.efficiency) { $jsonInput.Text = [string]$saved.efficiency }
    }
} catch {}

function Save-SourceConfig {
    try { @{ toolbox = $toolboxInput.Text.Trim(); jsonImport = $jsonInput.Text.Trim() } | ConvertTo-Json | Set-Content -LiteralPath $sourceConfig -Encoding UTF8 } catch {}
}

function Get-SourceArgs {
    Save-SourceConfig
    $json = $jsonInput.Text.Trim()
    $tool = $toolboxInput.Text.Trim()
    $argsList = @('-snapshot','none')
    if (-not [string]::IsNullOrWhiteSpace($json)) { $argsList += @('-efficiency',$json) }
    if (-not [string]::IsNullOrWhiteSpace($tool)) { $argsList += @('-toolbox',$tool) }
    if ([string]::IsNullOrWhiteSpace($json) -and [string]::IsNullOrWhiteSpace($tool)) { return @('-snapshot','auto') }
    return $argsList
}

$browse.Add_Click({
    $dlg = New-Object System.Windows.Forms.OpenFileDialog
    $dlg.Filter = 'JSON files (*.json)|*.json|All files (*.*)|*.*'
    $dlg.Title = 'Select IdleOn Toolbox / Efficiency account JSON'
    if ($dlg.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { $jsonInput.Text = $dlg.FileName; Save-SourceConfig }
})

$paste.Add_Click({
    try {
        if (-not [System.Windows.Clipboard]::ContainsText()) { throw 'Clipboard does not contain text.' }
        $text = [System.Windows.Clipboard]::GetText().Trim()
        if ([string]::IsNullOrWhiteSpace($text)) { throw 'Clipboard is empty.' }
        $null = $text | ConvertFrom-Json -ErrorAction Stop
        [IO.File]::WriteAllText($clipboardImport, $text, (New-Object System.Text.UTF8Encoding($false)))
        $jsonInput.Text = $clipboardImport
        Save-SourceConfig
        $output.Text = 'Clipboard JSON imported successfully.' + $nl + 'Click Source Check next.'
    } catch {
        $output.Text = 'CLIPBOARD IMPORT ERROR:' + $nl + $_.Exception.Message + $nl + $nl + 'Copy the complete Raw JSON from IdleOn Efficiency or a Toolbox account-data export, then try again.'
    }
})

$bar = New-Object System.Windows.Controls.WrapPanel
$bar.Margin = '0,0,0,12'
[System.Windows.Controls.Grid]::SetRow($bar,1)
$root.Children.Add($bar) | Out-Null

function New-AgentButton([string]$text) {
    $b = New-Object System.Windows.Controls.Button
    $b.Content = $text
    $b.Margin = '0,0,8,8'
    $b.Padding = '14,8'
    $b.MinWidth = 130
    $b.Background = '#252B31'
    $b.Foreground = '#F2F2F2'
    $b.BorderBrush = '#3A424A'
    return $b
}

$output = New-Object System.Windows.Controls.TextBox
$output.IsReadOnly = $true
$output.AcceptsReturn = $true
$output.TextWrapping = 'Wrap'
$output.VerticalScrollBarVisibility = 'Auto'
$output.HorizontalScrollBarVisibility = 'Auto'
$output.FontFamily = 'Consolas'
$output.FontSize = 13
$output.Background = '#171B1F'
$output.Foreground = '#E8ECEF'
$output.BorderBrush = '#30363D'
$output.Padding = '12'
$output.Text = 'Ready.' + $nl + '1) Detect Game' + $nl + '2) Select/import account data' + $nl + '3) Source Check' + $nl + '4) Assess Account' + $nl + $nl + 'Automation is locked in this test build.'
[System.Windows.Controls.Grid]::SetRow($output,2)
$root.Children.Add($output) | Out-Null

function Run-AgentCommand([string[]]$commandArgs) {
    try {
        $output.Text = 'Running...' + $nl + $exe + ' ' + (Join-Args $commandArgs)
        $window.Dispatcher.Invoke([action]{}, 'Background')
        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = $exe
        $psi.Arguments = Join-Args $commandArgs
        $psi.WorkingDirectory = $appDir
        $psi.UseShellExecute = $false
        $psi.RedirectStandardOutput = $true
        $psi.RedirectStandardError = $true
        $psi.CreateNoWindow = $true
        $p = New-Object System.Diagnostics.Process
        $p.StartInfo = $psi
        if (-not $p.Start()) { throw 'Process failed to start.' }
        $stdout = $p.StandardOutput.ReadToEnd()
        $stderr = $p.StandardError.ReadToEnd()
        $p.WaitForExit()
        $text = ($stdout + $stderr).Trim()
        if ([string]::IsNullOrWhiteSpace($text)) { $text = 'Command completed with exit code ' + $p.ExitCode + '.' }
        $output.Text = $text
        $output.ScrollToHome()
    } catch {
        $output.Text = 'COMMAND ERROR:' + $nl + $_.Exception.Message + $nl + $nl + $_.ScriptStackTrace
    }
}

$detect = New-AgentButton 'Detect Game'
$detect.Add_Click({ Run-AgentCommand @('detect') })
$bar.Children.Add($detect) | Out-Null

$doctor = New-AgentButton 'Source Check'
$doctor.Add_Click({ $sourceArgs = Get-SourceArgs; Run-AgentCommand (@('doctor','-config',$config) + $sourceArgs) })
$bar.Children.Add($doctor) | Out-Null

$assess = New-AgentButton 'Assess Account'
$assess.Background = '#21472E'
$assess.Add_Click({ $sourceArgs = Get-SourceArgs; Run-AgentCommand (@('assess','-config',$config) + $sourceArgs) })
$bar.Children.Add($assess) | Out-Null

$clear = New-AgentButton 'Clear Account Source'
$clear.Add_Click({ $toolboxInput.Text=''; $jsonInput.Text=''; Save-SourceConfig; $output.Text='Account source fields cleared. With both blank, Source Check will try the local companion snapshot path.' })
$bar.Children.Add($clear) | Out-Null

$locked = New-AgentButton 'Automation Locked'
$locked.IsEnabled = $false
$locked.Background = '#4A3030'
$locked.ToolTip = 'Automation comes after the account parser is validated against real account data.'
$bar.Children.Add($locked) | Out-Null

$window.Add_Closing({ Save-SourceConfig })
[void]$window.ShowDialog()
`

	cmd := exec.Command("powershell.exe", "-NoProfile", "-STA", "-WindowStyle", "Hidden", "-Command", script)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Env = append(os.Environ(), "IDLEON_AGENT_EXE="+exePath, "IDLEON_AGENT_CONFIG="+configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(output)); if msg != "" { return fmt.Errorf("native control panel: %w: %s", err, msg) }; return fmt.Errorf("native control panel: %w", err)
	}
	return nil
}

func showStartupError(err error) {
	message := "IdleOn Account Agent could not start:\n\n" + err.Error() + "\n\nOpen Command Prompt in this folder and run:\nIdleOn-Account-Agent.exe doctor -config automation.json"
	_ = exec.Command("powershell.exe", "-NoProfile", "-Command", "Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show($args[0], 'IdleOn Account Agent')", message).Run()
	fmt.Fprintln(os.Stderr, message)
}
