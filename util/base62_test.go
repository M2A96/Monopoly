// util/base62_test.go
package util

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDToBase62(t *testing.T) {
	tests := []struct {
		name string
		uuid uuid.UUID
		want string
	}{
		{
			name: "Zero UUID",
			uuid: uuid.UUID{},
			want: "0",
		},
		{
			name: "Max UUID",
			uuid: uuid.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			want: "7n42DGM5Tflk9n8mt7Fhc7",
		},
		{
			name: "Random UUID 1",
			uuid: uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
			want: "3odNFKDXqcJgvzf0aHhbQz",
		},
		{
			name: "Random UUID 2",
			uuid: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			want: "2AJyxJiVHVrNMDDkiEIW0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UUIDToBase62(tt.uuid)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBase62ToUUID(t *testing.T) {
	tests := []struct {
		name    string
		base62  string
		want    uuid.UUID
		wantErr bool
	}{
		{
			name:    "Zero UUID",
			base62:  "0",
			want:    uuid.UUID{},
			wantErr: false,
		},
		{
			name:    "Max UUID",
			base62:  "7n42DGM5Tflk9n8mt7Fhc7",
			want:    uuid.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantErr: false,
		},
		{
			name:    "Random UUID 1",
			base62:  "3odNFKDXqcJgvzf0aHhbQz",
			want:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
			wantErr: false,
		},
		{
			name:    "Random UUID 2",
			base62:  "2AJyxJiVHVrNMDDkiEIW0",
			want:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base62ToUUID(tt.base62)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that converting a UUID to base62 and back results in the original UUID
	tests := []struct {
		name string
		uuid uuid.UUID
	}{
		{
			name: "Zero UUID",
			uuid: uuid.UUID{},
		},
		{
			name: "Max UUID",
			uuid: uuid.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name: "Random UUID 1",
			uuid: uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
		},
		{
			name: "Random UUID 2",
			uuid: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name: "Random UUID 3",
			uuid: uuid.New(), // Generate a new random UUID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base62 := UUIDToBase62(tt.uuid)
			gotUUID, err := Base62ToUUID(base62)
			assert.NoError(t, err)
			assert.Equal(t, tt.uuid, gotUUID)
		})
	}
}

func BenchmarkUUIDToBase62(b *testing.B) {
	id := uuid.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UUIDToBase62(id)
	}
}

func BenchmarkBase62ToUUID(b *testing.B) {
	id := uuid.New()
	base62 := UUIDToBase62(id)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Base62ToUUID(base62)
	}
}
