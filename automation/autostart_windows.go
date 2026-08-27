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
	if len(os.Args) > 1 {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		showStartupError(err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exe)
	_ = os.Chdir(exeDir)

	const configPath = "automation.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if saveErr := saveConfig(configPath, defaultConfig()); saveErr != nil {
			showStartupError(fmt.Errorf("create starter config: %w", saveErr))
			os.Exit(1)
		}
	}

	if err := launchNativeControlPanel(exe, filepath.Join(exeDir, configPath)); err != nil {
		showStartupError(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func launchNativeControlPanel(exePath, configPath string) error {
	const script = `
Add-Type -AssemblyName PresentationFramework
Add-Type -AssemblyName PresentationCore
Add-Type -AssemblyName WindowsBase

$exe = $env:IDLEON_AGENT_EXE
$config = $env:IDLEON_AGENT_CONFIG
$script:agentProcess = $null
$nl = [Environment]::NewLine

if ([string]::IsNullOrWhiteSpace($exe) -or -not (Test-Path -LiteralPath $exe)) {
    [System.Windows.MessageBox]::Show('Backend executable path is missing or invalid: ' + $exe, 'IdleOn Account Agent') | Out-Null
    exit 2
}

function Quote-Arg([string]$s) {
    if ($null -eq $s) { return '""' }
    return '"' + ($s -replace '"','\"') + '"'
}

function Join-Args([string[]]$items) {
    return (($items | ForEach-Object { Quote-Arg $_ }) -join ' ')
}

$window = New-Object System.Windows.Window
$window.Title = 'IdleOn Account Agent'
$window.Width = 980
$window.Height = 720
$window.MinWidth = 760
$window.MinHeight = 560
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
$title.FontSize = 26
$title.FontWeight = 'Bold'
$title.Foreground = '#F2F2F2'
$header.Children.Add($title) | Out-Null

$subtitle = New-Object System.Windows.Controls.TextBlock
$subtitle.Text = 'Whole-account assessment and calibrated automation. F12 is the emergency stop.'
$subtitle.Margin = '0,4,0,12'
$subtitle.Foreground = '#AAB4BE'
$header.Children.Add($subtitle) | Out-Null

$bar = New-Object System.Windows.Controls.WrapPanel
$bar.Margin = '0,0,0,12'
[System.Windows.Controls.Grid]::SetRow($bar,1)
$root.Children.Add($bar) | Out-Null

function New-AgentButton([string]$text) {
    $b = New-Object System.Windows.Controls.Button
    $b.Content = $text
    $b.Margin = '0,0,8,8'
    $b.Padding = '14,8'
    $b.MinWidth = 120
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
$output.FontFamily = 'Consolas'
$output.FontSize = 13
$output.Background = '#171B1F'
$output.Foreground = '#E8ECEF'
$output.BorderBrush = '#30363D'
$output.Padding = '12'
$output.Text = 'Ready. Start IdleOn, then click Detect Game.' + $nl + 'Backend: ' + $exe
[System.Windows.Controls.Grid]::SetRow($output,2)
$root.Children.Add($output) | Out-Null

function Run-AgentCommand([string[]]$commandArgs) {
    try {
        $output.Text = 'Running: ' + $exe + ' ' + (Join-Args $commandArgs)
        $window.Dispatcher.Invoke([action]{}, 'Background')

        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = $exe
        $psi.Arguments = Join-Args $commandArgs
        $psi.WorkingDirectory = Split-Path -Parent $exe
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
        $code = $p.ExitCode
        $text = ($stdout + $stderr).Trim()
        if ([string]::IsNullOrWhiteSpace($text)) {
            $text = 'Command completed with exit code ' + $code + '.'
        }
        $output.Text = $text
        $output.ScrollToEnd()
    } catch {
        $output.Text = 'COMMAND ERROR:' + $nl + $_.Exception.Message + $nl + $nl + $_.ScriptStackTrace
        $output.ScrollToEnd()
    }
}

$detect = New-AgentButton 'Detect Game'
$detect.Add_Click({ Run-AgentCommand @('detect') })
$bar.Children.Add($detect) | Out-Null

$doctor = New-AgentButton 'Doctor'
$doctor.Add_Click({ Run-AgentCommand @('doctor','-config',$config,'-snapshot','auto') })
$bar.Children.Add($doctor) | Out-Null

$assess = New-AgentButton 'Assess Account'
$assess.Add_Click({ Run-AgentCommand @('assess','-config',$config,'-snapshot','auto') })
$bar.Children.Add($assess) | Out-Null

$calibrate = New-AgentButton 'Calibrate'
$calibrate.Add_Click({
    $output.Text = 'Calibration started. Keep IdleOn visible. Move to each requested point and press F8. F12 aborts.'
    Run-AgentCommand @('calibrate','-config',$config)
})
$bar.Children.Add($calibrate) | Out-Null

$start = New-AgentButton 'Start Automation'
$start.Background = '#21472E'
$start.Add_Click({
    try {
        if ($script:agentProcess -and -not $script:agentProcess.HasExited) {
            $output.Text = 'Automation is already running. F12 is the emergency stop.'
            return
        }
        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = $exe
        $psi.Arguments = Join-Args @('agent','-config',$config,'-snapshot','auto','-execute','-foreground-only=false')
        $psi.WorkingDirectory = Split-Path -Parent $exe
        $psi.UseShellExecute = $false
        $psi.CreateNoWindow = $true
        $script:agentProcess = [System.Diagnostics.Process]::Start($psi)
        if ($null -eq $script:agentProcess) { throw 'Agent process failed to start.' }
        $output.Text = 'Automation process started (PID ' + $script:agentProcess.Id + ').' + $nl + 'The agent will assess in the background. Press F12 to emergency-stop an active routine.'
    } catch {
        $output.Text = 'START ERROR:' + $nl + $_.Exception.Message
    }
})
$bar.Children.Add($start) | Out-Null

$stop = New-AgentButton 'Stop Agent'
$stop.Background = '#53282B'
$stop.Add_Click({
    if ($script:agentProcess -and -not $script:agentProcess.HasExited) {
        try {
            $script:agentProcess.Kill()
            $script:agentProcess.WaitForExit(3000) | Out-Null
            $output.Text = 'Agent process stopped.'
        } catch {
            $output.Text = 'STOP ERROR: ' + $_.Exception.Message
        }
    } else {
        $output.Text = 'No automation process is currently running.'
    }
})
$bar.Children.Add($stop) | Out-Null

$web = New-AgentButton 'Web Diagnostics'
$web.Add_Click({
    $output.Text = 'Web diagnostics are optional only. Normal operation stays in this Windows app.' + $nl + 'CLI: IdleOn-Account-Agent.exe serve -config automation.json'
})
$bar.Children.Add($web) | Out-Null

$window.Add_Closing({
    if ($script:agentProcess -and -not $script:agentProcess.HasExited) {
        try { $script:agentProcess.Kill() } catch {}
    }
})

[void]$window.ShowDialog()
`

	cmd := exec.Command("powershell.exe", "-NoProfile", "-STA", "-WindowStyle", "Hidden", "-Command", script)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Env = append(os.Environ(),
		"IDLEON_AGENT_EXE="+exePath,
		"IDLEON_AGENT_CONFIG="+configPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			return fmt.Errorf("native control panel: %w: %s", err, msg)
		}
		return fmt.Errorf("native control panel: %w", err)
	}
	return nil
}

func showStartupError(err error) {
	message := "IdleOn Account Agent could not start:\n\n" + err.Error() + "\n\nOpen Command Prompt in this folder and run:\nIdleOn-Account-Agent.exe doctor -config automation.json"
	_ = exec.Command("powershell.exe", "-NoProfile", "-Command", "Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show($args[0], 'IdleOn Account Agent')", message).Run()
	fmt.Fprintln(os.Stderr, message)
}
