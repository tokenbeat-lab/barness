package ai

// ImageModel describes an image operation's model. Its identity is
// (image, Provider, API, ID); it does not share chat limits or compatibility.
type ImageModel struct {
	Provider     ProviderID        `json:"provider"`
	API          API               `json:"api"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Input        []Modality        `json:"input"`
	Output       []Modality        `json:"output"`
	Capabilities ImageCapabilities `json:"capabilities"`
	Pricing      ImagePricing      `json:"pricing"`
}

// ImageCapabilities are model-specific limits and supported image options.
// Empty option lists mean those options must be omitted.
type ImageCapabilities struct {
	MaxReferenceImages int      `json:"maxReferenceImages"`
	MaxOutputImages    int      `json:"maxOutputImages"`
	Sizes              []string `json:"sizes,omitempty"`
	// CustomSizes enables OpenAI's bounded WIDTHxHEIGHT dimensions in
	// addition to Sizes. Qualities explicitly lists supported quality values.
	CustomSizes           bool     `json:"customSizes,omitempty"`
	Qualities             []string `json:"qualities,omitempty"`
	ImageSizes            []string `json:"imageSizes,omitempty"`
	AspectRatios          []string `json:"aspectRatios,omitempty"`
	Mask                  bool     `json:"mask,omitempty"`
	TransparentBackground bool     `json:"transparentBackground,omitempty"`
	InputFidelity         []string `json:"inputFidelity,omitempty"`
}

// ImagePricing contains US dollars per million tokens, split by modality.
// Every declared input/output modality requires a value; explicit zero is a
// free rate, while null or unset is missing. Cache-read rates are optional.
type ImagePricing struct {
	InputText      Nullable[float64] `json:"inputText,omitzero"`
	InputImage     Nullable[float64] `json:"inputImage,omitzero"`
	OutputText     Nullable[float64] `json:"outputText,omitzero"`
	OutputImage    Nullable[float64] `json:"outputImage,omitzero"`
	CacheReadText  Nullable[float64] `json:"cacheReadText,omitzero"`
	CacheReadImage Nullable[float64] `json:"cacheReadImage,omitzero"`
	// Source names the official pricing URL and the date the rates were read.
	Source string `json:"source"`
}

// ClassifierModel describes a classifier operation's model, located by
// (classifier, Provider, API, ID). It has no chat generation settings.
type ClassifierModel struct {
	Provider      ProviderID             `json:"provider"`
	API           API                    `json:"api"`
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	ContextWindow int                    `json:"contextWindow"`
	Capabilities  ClassifierCapabilities `json:"capabilities"`
	Cost          ModelCost              `json:"cost"`
}

// ClassifierQuestionKind is a classifier's supported question category.
// Its JSON encoding remains a string. Values from host configuration still
// require catalog validation; an empty or unknown kind is refused by NewClient.
type ClassifierQuestionKind string

const (
	ClassifierQuestionChoice ClassifierQuestionKind = "choice"
	ClassifierQuestionScore  ClassifierQuestionKind = "score"
	ClassifierQuestionBool   ClassifierQuestionKind = "bool"
)

// ClassifierCapabilities names supported question kinds (choice, score,
// bool). Choice requires MaxChoices >= 2; score requires an inclusive level
// range with minimum >= 2. Limits for unsupported kinds must be zero.
type ClassifierCapabilities struct {
	Kinds          []ClassifierQuestionKind `json:"kinds"`
	MaxChoices     int                      `json:"maxChoices"`
	MinScoreLevels int                      `json:"minScoreLevels"`
	MaxScoreLevels int                      `json:"maxScoreLevels"`
}
