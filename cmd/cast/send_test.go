package main

import (
	"testing"
)

func TestProcessNewlines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "sem quebras de linha",
			input:    "Mensagem simples",
			expected: "Mensagem simples",
		},
		{
			name:     "quebra de linha simples",
			input:    "Linha 1\\nLinha 2",
			expected: "Linha 1\nLinha 2",
		},
		{
			name:     "linha em branco (duas quebras)",
			input:    "Parágrafo 1\\n\\nParágrafo 2",
			expected: "Parágrafo 1\n\nParágrafo 2",
		},
		{
			name:     "múltiplas quebras de linha",
			input:    "A\\nB\\nC",
			expected: "A\nB\nC",
		},
		{
			name:     "quebra simples e linha em branco",
			input:    "Linha 1\\nLinha 2\\n\\nLinha 3",
			expected: "Linha 1\nLinha 2\n\nLinha 3",
		},
		{
			name:     "texto com backslash n literal (não processado)",
			input:    "Texto com \\n no meio",
			expected: "Texto com \n no meio",
		},
		{
			name:     "quebra no início",
			input:    "\\nTexto após quebra",
			expected: "\nTexto após quebra",
		},
		{
			name:     "quebra no final",
			input:    "Texto antes quebra\\n",
			expected: "Texto antes quebra\n",
		},
		{
			name:     "linha em branco no início",
			input:    "\\n\\nTexto após linha em branco",
			expected: "\n\nTexto após linha em branco",
		},
		{
			name:     "linha em branco no final",
			input:    "Texto antes linha em branco\\n\\n",
			expected: "Texto antes linha em branco\n\n",
		},
		{
			name:     "múltiplas linhas em branco",
			input:    "A\\n\\nB\\n\\nC",
			expected: "A\n\nB\n\nC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processNewlines(tt.input)
			if result != tt.expected {
				t.Errorf("processNewlines(%q) = %q, esperado %q", tt.input, result, tt.expected)
			}
		})
	}
}
