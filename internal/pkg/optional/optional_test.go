package optional

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  Optional[string]
		want string
	}{
		{
			name: "string value",
			got:  Some("hello"),
			want: "hello",
		},
		{
			name: "empty string",
			got:  Some(""),
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.True(t, tt.got.Defined)
			require.NotNil(t, tt.got.Value)
			assert.Equal(t, tt.want, *tt.got.Value)
		})
	}

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		got := Some(42)
		require.True(t, got.Defined)
		require.NotNil(t, got.Value)
		assert.Equal(t, 42, *got.Value)
	})

	t.Run("bool value", func(t *testing.T) {
		t.Parallel()
		got := Some(true)
		require.True(t, got.Defined)
		require.NotNil(t, got.Value)
		assert.True(t, *got.Value)
	})
}

func TestNone(t *testing.T) {
	t.Parallel()

	got := None[string]()
	assert.False(t, got.Defined)
	assert.Nil(t, got.Value)

	gotInt := None[int]()
	assert.False(t, gotInt.Defined)
	assert.Nil(t, gotInt.Value)
}

func TestNull(t *testing.T) {
	t.Parallel()

	got := Null[string]()
	assert.True(t, got.Defined)
	assert.Nil(t, got.Value)

	gotInt := Null[int]()
	assert.True(t, gotInt.Defined)
	assert.Nil(t, gotInt.Value)
}

func TestOptional_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opt  any
		want string
	}{
		{
			name: "undefined string",
			opt:  None[string](),
			want: "null",
		},
		{
			name: "explicit null string",
			opt:  Null[string](),
			want: `{"null_protected_value":null}`,
		},
		{
			name: "defined string",
			opt:  Some("hello"),
			want: `"hello"`,
		},
		{
			name: "defined empty string",
			opt:  Some(""),
			want: `""`,
		},
		{
			name: "undefined int",
			opt:  None[int](),
			want: "null",
		},
		{
			name: "explicit null int",
			opt:  Null[int](),
			want: `{"null_protected_value":null}`,
		},
		{
			name: "defined int",
			opt:  Some(42),
			want: "42",
		},
		{
			name: "defined bool",
			opt:  Some(true),
			want: "true",
		},
		{
			name: "defined slice",
			opt:  Some([]string{"a", "b"}),
			want: `["a","b"]`,
		},
		{
			name: "explicit null pointer type",
			opt:  Null[*string](),
			want: `{"null_protected_value":null}`,
		},
		{
			name: "defined nil pointer value",
			opt:  Some[*string](nil),
			want: "null",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := json.Marshal(tt.opt)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestOptional_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("string value", func(t *testing.T) {
		t.Parallel()
		var opt Optional[string]
		require.NoError(t, json.Unmarshal([]byte(`"hello"`), &opt))
		require.True(t, opt.Defined)
		require.NotNil(t, opt.Value)
		assert.Equal(t, "hello", *opt.Value)
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		var opt Optional[string]
		require.NoError(t, json.Unmarshal([]byte(`""`), &opt))
		require.True(t, opt.Defined)
		require.NotNil(t, opt.Value)
		assert.Equal(t, "", *opt.Value)
	})

	t.Run("null becomes explicit null", func(t *testing.T) {
		t.Parallel()
		var opt Optional[string]
		require.NoError(t, json.Unmarshal([]byte(`null`), &opt))
		assert.True(t, opt.Defined)
		assert.Nil(t, opt.Value)
	})

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		var opt Optional[int]
		require.NoError(t, json.Unmarshal([]byte(`7`), &opt))
		require.True(t, opt.Defined)
		require.NotNil(t, opt.Value)
		assert.Equal(t, 7, *opt.Value)
	})

	t.Run("bool value", func(t *testing.T) {
		t.Parallel()
		var opt Optional[bool]
		require.NoError(t, json.Unmarshal([]byte(`false`), &opt))
		require.True(t, opt.Defined)
		require.NotNil(t, opt.Value)
		assert.False(t, *opt.Value)
	})

	t.Run("slice value", func(t *testing.T) {
		t.Parallel()
		var opt Optional[[]string]
		require.NoError(t, json.Unmarshal([]byte(`["a","b"]`), &opt))
		require.True(t, opt.Defined)
		require.NotNil(t, opt.Value)
		assert.Equal(t, []string{"a", "b"}, *opt.Value)
	})

	t.Run("invalid payload", func(t *testing.T) {
		t.Parallel()
		var opt Optional[int]
		err := json.Unmarshal([]byte(`"not-an-int"`), &opt)
		require.Error(t, err)
		assert.True(t, opt.Defined, "Defined is set before unmarshaling the value")
	})
}

func TestOptional_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("some string round-trips", func(t *testing.T) {
		t.Parallel()
		original := Some("hello")

		data, err := json.Marshal(original)
		require.NoError(t, err)
		assert.JSONEq(t, `"hello"`, string(data))

		var got Optional[string]
		require.NoError(t, json.Unmarshal(data, &got))
		require.True(t, got.Defined)
		require.NotNil(t, got.Value)
		assert.Equal(t, "hello", *got.Value)
	})

	t.Run("json null unmarshals as explicit null", func(t *testing.T) {
		t.Parallel()
		var got Optional[string]
		require.NoError(t, json.Unmarshal([]byte("null"), &got))
		assert.True(t, got.Defined)
		assert.Nil(t, got.Value)
	})

	t.Run("null marshals as protected object", func(t *testing.T) {
		t.Parallel()
		data, err := json.Marshal(Null[string]())
		require.NoError(t, err)
		assert.JSONEq(t, `{"null_protected_value":null}`, string(data))
	})
}

func TestOptional_InStruct(t *testing.T) {
	t.Parallel()

	type patch struct {
		Title Optional[string] `json:"title"`
		Count Optional[int]    `json:"count"`
	}

	t.Run("marshal defined fields", func(t *testing.T) {
		t.Parallel()
		p := patch{
			Title: Some("updated"),
			Count: Some(3),
		}
		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.JSONEq(t, `{"title":"updated","count":3}`, string(data))
	})

	t.Run("marshal explicit null field", func(t *testing.T) {
		t.Parallel()
		p := patch{
			Title: Null[string](),
			Count: None[int](),
		}
		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.JSONEq(t, `{"title":{"null_protected_value":null},"count":null}`, string(data))
	})

	t.Run("unmarshal sets defined", func(t *testing.T) {
		t.Parallel()
		var p patch
		require.NoError(t, json.Unmarshal([]byte(`{"title":"new","count":null}`), &p))
		require.True(t, p.Title.Defined)
		require.NotNil(t, p.Title.Value)
		assert.Equal(t, "new", *p.Title.Value)
		assert.True(t, p.Count.Defined)
		assert.Nil(t, p.Count.Value)
	})

	t.Run("absent field stays undefined", func(t *testing.T) {
		t.Parallel()
		var p patch
		require.NoError(t, json.Unmarshal([]byte(`{"title":"only-title"}`), &p))
		require.True(t, p.Title.Defined)
		assert.False(t, p.Count.Defined)
		assert.Nil(t, p.Count.Value)
	})
}

func TestOptional_Time(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	opt := Some(ts)

	data, err := json.Marshal(opt)
	require.NoError(t, err)

	var got Optional[time.Time]
	require.NoError(t, json.Unmarshal(data, &got))
	require.True(t, got.Defined)
	require.NotNil(t, got.Value)
	assert.True(t, ts.Equal(*got.Value))
}

func TestNullProtectedValueTag(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "null_protected_value", NullProtectedValueTag)
}
