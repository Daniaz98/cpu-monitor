package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID    int32
	Name   string
	CPU    float64
	Memory uint64
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func getTopProcess() []ProcessInfo {
	processes, _ := process.Processes()
	var processList []ProcessInfo

	for _, p := range processes {
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()

		if cpuPercent > 0 {
			processList = append(processList, ProcessInfo{
				PID:    p.Pid,
				Name:   name,
				CPU:    cpuPercent,
				Memory: memInfo.RSS, //memória usada (Resident Set Size)
			})
		}
	}

	// Ordena os processos primeiro por CPU, depois por Memória
	sort.Slice(processList, func(i, j int) bool {
		return processList[i].CPU > processList[j].CPU
	})

	// Retorna apenas os 5 principais
	if len(processList) > 5 {
		return processList[:5]
	}
	return processList
}

func monitor() {
	for {
		clearScreen()

		cpuUsage, _ := cpu.Percent(time.Second, false)

		memInfo, _ := mem.VirtualMemory()

		fmt.Printf("Uso da CPU: %.2f%%\n", cpuUsage[0])
		fmt.Printf("Memória: %.2fGB / %.2fGB (%.2f%%)\n", float64(memInfo.Used)/(1024*1024*1024), float64(memInfo.Total)/(1024*1024*1024), memInfo.UsedPercent)

		if cpuUsage[0] > 80 {
			fmt.Println("ALERTA: Uso da CPU acima de 80%!")
		}
		if memInfo.UsedPercent > 85 {
			fmt.Println("ALERTA: Uso da memória acima de 85%!")
		}

		fmt.Println("\n Processos que mais consomem CPU:")
		topProcesses := getTopProcess()
		for _, p := range topProcesses {
			fmt.Printf("PID: %d | %s | CPU: %.2f%% | Memória: %.2fMB\n",
				p.PID, p.Name, p.CPU, float64(p.Memory)/(1024*1024))
		}

		time.Sleep(10 * time.Second)
	}
}

func reportsCSV(processes []ProcessInfo, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"PID", "Nome", "Uso de CPU (%)", "Uso de Memória (bytes)"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, p := range processes {
		line := []string{
			strconv.Itoa(int(p.PID)),
			p.Name,
			fmt.Sprintf("%.2f", p.CPU),
			formatMemory(p.Memory),
		}
		writer.Write(line)
	}

	fmt.Println("Relatório gerado:", filename)
	return nil

}

func formatMemory(mem uint64) string {
	const MB = 1024 * 1024
	const GB = 1024 * 1024 * 1024

	if mem >= GB {
		return fmt.Sprintf("%.2f GB", float64(mem)/float64(GB))
	}
	return fmt.Sprintf("%.2f MB", float64(mem)/float64(MB))
}

func main() {
	fmt.Println("Iniciando monitor de CPU e Memória...(Ctrl + C para sair)")
	processes := getTopProcess()
	err := reportsCSV(processes, "relatorio_processos.csv")
	if err != nil {
		fmt.Println("Erro ao gerar relatório", err)
	}
	monitor()
}
