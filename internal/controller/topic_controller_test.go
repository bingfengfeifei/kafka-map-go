package controller

import (
	"testing"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
)

func TestParseClusterIDValue(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    uint
		wantErr bool
	}{
		{
			name:  "string cluster id",
			value: "1",
			want:  1,
		},
		{
			name:  "numeric cluster id from JSON",
			value: float64(2),
			want:  2,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "fractional JSON number",
			value:   float64(1.5),
			wantErr: true,
		},
		{
			name:    "missing cluster id",
			value:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseClusterIDValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseClusterIDValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("parseClusterIDValue() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizeCreateTopicRequestUsesNumPartitions(t *testing.T) {
	req := &dto.CreateTopicRequest{
		Name:              "orders",
		NumPartitions:     3,
		ReplicationFactor: 1,
	}

	if err := normalizeCreateTopicRequest(req); err != nil {
		t.Fatalf("normalizeCreateTopicRequest() error = %v", err)
	}
	if req.Partitions != 3 {
		t.Fatalf("Partitions = %d, want 3", req.Partitions)
	}
}

func TestNormalizeCreateTopicRequestRejectsMissingPartitions(t *testing.T) {
	req := &dto.CreateTopicRequest{
		Name:              "orders",
		ReplicationFactor: 1,
	}

	if err := normalizeCreateTopicRequest(req); err == nil {
		t.Fatal("normalizeCreateTopicRequest() error = nil, want error")
	}
}
