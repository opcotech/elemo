package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLanguageCode_String(t *testing.T) {
	tests := []struct {
		name string
		s    Language
		want string
	}{
		{"aa", LanguageAA, "aa"},
		{"ab", LanguageAB, "ab"},
		{"ae", LanguageAE, "ae"},
		{"af", LanguageAF, "af"},
		{"ak", LanguageAK, "ak"},
		{"am", LanguageAM, "am"},
		{"an", LanguageAN, "an"},
		{"ar", LanguageAR, "ar"},
		{"as", LanguageAS, "as"},
		{"av", LanguageAV, "av"},
		{"ay", LanguageAY, "ay"},
		{"az", LanguageAZ, "az"},
		{"ba", LanguageBA, "ba"},
		{"be", LanguageBE, "be"},
		{"bg", LanguageBG, "bg"},
		{"bh", LanguageBH, "bh"},
		{"bi", LanguageBI, "bi"},
		{"bm", LanguageBM, "bm"},
		{"bn", LanguageBN, "bn"},
		{"bo", LanguageBO, "bo"},
		{"br", LanguageBR, "br"},
		{"bs", LanguageBS, "bs"},
		{"ca", LanguageCA, "ca"},
		{"ce", LanguageCE, "ce"},
		{"ch", LanguageCH, "ch"},
		{"co", LanguageCO, "co"},
		{"cr", LanguageCR, "cr"},
		{"cs", LanguageCS, "cs"},
		{"cu", LanguageCU, "cu"},
		{"cv", LanguageCV, "cv"},
		{"cy", LanguageCY, "cy"},
		{"da", LanguageDA, "da"},
		{"de", LanguageDE, "de"},
		{"dv", LanguageDV, "dv"},
		{"dz", LanguageDZ, "dz"},
		{"ee", LanguageEE, "ee"},
		{"el", LanguageEL, "el"},
		{"en", LanguageEN, "en"},
		{"eo", LanguageEO, "eo"},
		{"es", LanguageES, "es"},
		{"et", LanguageET, "et"},
		{"eu", LanguageEU, "eu"},
		{"fa", LanguageFA, "fa"},
		{"ff", LanguageFF, "ff"},
		{"fi", LanguageFI, "fi"},
		{"fj", LanguageFJ, "fj"},
		{"fo", LanguageFO, "fo"},
		{"fr", LanguageFR, "fr"},
		{"fy", LanguageFY, "fy"},
		{"ga", LanguageGA, "ga"},
		{"gd", LanguageGD, "gd"},
		{"gl", LanguageGL, "gl"},
		{"gn", LanguageGN, "gn"},
		{"gu", LanguageGU, "gu"},
		{"gv", LanguageGV, "gv"},
		{"ha", LanguageHA, "ha"},
		{"he", LanguageHE, "he"},
		{"hi", LanguageHI, "hi"},
		{"ho", LanguageHO, "ho"},
		{"hr", LanguageHR, "hr"},
		{"ht", LanguageHT, "ht"},
		{"hu", LanguageHU, "hu"},
		{"hy", LanguageHY, "hy"},
		{"hz", LanguageHZ, "hz"},
		{"ia", LanguageIA, "ia"},
		{"id", LanguageID, "id"},
		{"ie", LanguageIE, "ie"},
		{"ig", LanguageIG, "ig"},
		{"ii", LanguageII, "ii"},
		{"ik", LanguageIK, "ik"},
		{"io", LanguageIO, "io"},
		{"is", LanguageIS, "is"},
		{"it", LanguageIT, "it"},
		{"iu", LanguageIU, "iu"},
		{"ja", LanguageJA, "ja"},
		{"jv", LanguageJV, "jv"},
		{"ka", LanguageKA, "ka"},
		{"kg", LanguageKG, "kg"},
		{"ki", LanguageKI, "ki"},
		{"kj", LanguageKJ, "kj"},
		{"kk", LanguageKK, "kk"},
		{"kl", LanguageKL, "kl"},
		{"km", LanguageKM, "km"},
		{"kn", LanguageKN, "kn"},
		{"ko", LanguageKO, "ko"},
		{"kr", LanguageKR, "kr"},
		{"ks", LanguageKS, "ks"},
		{"ku", LanguageKU, "ku"},
		{"kv", LanguageKV, "kv"},
		{"kw", LanguageKW, "kw"},
		{"ky", LanguageKY, "ky"},
		{"la", LanguageLA, "la"},
		{"lb", LanguageLB, "lb"},
		{"lg", LanguageLG, "lg"},
		{"li", LanguageLI, "li"},
		{"ln", LanguageLN, "ln"},
		{"lo", LanguageLO, "lo"},
		{"lt", LanguageLT, "lt"},
		{"lu", LanguageLU, "lu"},
		{"lv", LanguageLV, "lv"},
		{"mg", LanguageMG, "mg"},
		{"mh", LanguageMH, "mh"},
		{"mi", LanguageMI, "mi"},
		{"mk", LanguageMK, "mk"},
		{"ml", LanguageML, "ml"},
		{"mn", LanguageMN, "mn"},
		{"mr", LanguageMR, "mr"},
		{"ms", LanguageMS, "ms"},
		{"mt", LanguageMT, "mt"},
		{"my", LanguageMY, "my"},
		{"na", LanguageNA, "na"},
		{"nb", LanguageNB, "nb"},
		{"nd", LanguageND, "nd"},
		{"ne", LanguageNE, "ne"},
		{"ng", LanguageNG, "ng"},
		{"nl", LanguageNL, "nl"},
		{"nn", LanguageNN, "nn"},
		{"no", LanguageNO, "no"},
		{"nr", LanguageNR, "nr"},
		{"nv", LanguageNV, "nv"},
		{"ny", LanguageNY, "ny"},
		{"oc", LanguageOC, "oc"},
		{"oj", LanguageOJ, "oj"},
		{"om", LanguageOM, "om"},
		{"or", LanguageOR, "or"},
		{"os", LanguageOS, "os"},
		{"pa", LanguagePA, "pa"},
		{"pi", LanguagePI, "pi"},
		{"pl", LanguagePL, "pl"},
		{"ps", LanguagePS, "ps"},
		{"pt", LanguagePT, "pt"},
		{"qu", LanguageQU, "qu"},
		{"rm", LanguageRM, "rm"},
		{"rn", LanguageRN, "rn"},
		{"ro", LanguageRO, "ro"},
		{"ru", LanguageRU, "ru"},
		{"rw", LanguageRW, "rw"},
		{"sa", LanguageSA, "sa"},
		{"sc", LanguageSC, "sc"},
		{"sd", LanguageSD, "sd"},
		{"se", LanguageSE, "se"},
		{"sg", LanguageSG, "sg"},
		{"si", LanguageSI, "si"},
		{"sk", LanguageSK, "sk"},
		{"sl", LanguageSL, "sl"},
		{"sm", LanguageSM, "sm"},
		{"sn", LanguageSN, "sn"},
		{"so", LanguageSO, "so"},
		{"sq", LanguageSQ, "sq"},
		{"sr", LanguageSR, "sr"},
		{"ss", LanguageSS, "ss"},
		{"st", LanguageST, "st"},
		{"su", LanguageSU, "su"},
		{"sv", LanguageSV, "sv"},
		{"sw", LanguageSW, "sw"},
		{"ta", LanguageTA, "ta"},
		{"te", LanguageTE, "te"},
		{"tg", LanguageTG, "tg"},
		{"th", LanguageTH, "th"},
		{"ti", LanguageTI, "ti"},
		{"tk", LanguageTK, "tk"},
		{"tl", LanguageTL, "tl"},
		{"tn", LanguageTN, "tn"},
		{"to", LanguageTO, "to"},
		{"tr", LanguageTR, "tr"},
		{"ts", LanguageTS, "ts"},
		{"tt", LanguageTT, "tt"},
		{"tw", LanguageTW, "tw"},
		{"ty", LanguageTY, "ty"},
		{"ug", LanguageUG, "ug"},
		{"uk", LanguageUK, "uk"},
		{"ur", LanguageUR, "ur"},
		{"uz", LanguageUZ, "uz"},
		{"ve", LanguageVE, "ve"},
		{"vi", LanguageVI, "vi"},
		{"vo", LanguageVO, "vo"},
		{"wa", LanguageWA, "wa"},
		{"wo", LanguageWO, "wo"},
		{"xh", LanguageXH, "xh"},
		{"yi", LanguageYI, "yi"},
		{"yo", LanguageYO, "yo"},
		{"za", LanguageZA, "za"},
		{"zh", LanguageZH, "zh"},
		{"zu", LanguageZU, "zu"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestLanguage_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       Language
		want    []byte
		wantErr bool
	}{
		{"aa", LanguageAA, []byte("aa"), false},
		{"ab", LanguageAB, []byte("ab"), false},
		{"code high", Language(255), nil, true},
		{"code low", Language(0), nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.MarshalText()
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestLanguage_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *Language
		text    []byte
		want    Language
		wantErr bool
	}{
		{"aa", new(Language), []byte("aa"), LanguageAA, false},
		{"ab", new(Language), []byte("ab"), LanguageAB, false},
		{"code high", new(Language), []byte("255"), Language(0), true},
		{"code low", new(Language), []byte("0"), Language(0), true},
		{"code invalid", new(Language), []byte("invalid"), Language(0), true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.s.UnmarshalText([]byte(tt.name))
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}
