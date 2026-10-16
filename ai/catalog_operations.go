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
	// CustomSizes authorizes WIDTHxHEIGHT using sourced model constraints.
	// Numeric Sizes also obey these constraints when present.
	CustomSizes           *ImageSizeConstraints `json:"customSizes,omitempty"`
	Qualities             []string              `json:"qualities,omitempty"`
	ImageSizes            []string              `json:"imageSizes,omitempty"`
	AspectRatios          []string              `json:"aspectRatios,omitempty"`
	Mask                  bool                  `json:"mask,omitempty"`
	TransparentBackground bool                  `json:"transparentBackground,omitempty"`
	InputFidelity         []string              `json:"inputFidelity,omitempty"`
}

// ImageSizeConstraints describes a model's arbitrary output dimensions.
// Source records the capability evidence and retrieval date; no adapter
// infers limits from a model ID. Pixels are width multiplied by height.
type ImageSizeConstraints struct {
	EdgeMultiple   int    `json:"edgeMultiple"`
	MaxAspectRatio int    `json:"maxAspectRatio"`
	MaxEdge        int    `json:"maxEdge"`
	MinPixels      int64  `json:"minPixels"`
	MaxPixels      int64  `json:"maxPixels"`
	Source         string `json:"source"`
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
	Provider      ProviderID `json:"provider"`
	API           API        `json:"api"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ContextWindow int        `json:"contextWindow"`
	// Omitted input retains the legacy text-only contract.
	Input        []Modality             `json:"input,omitempty"`
	Capabilities ClassifierCapabilities `json:"capabilities"`
	Cost         ModelCost              `json:"cost"`
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
// bool). Omitted MinChoices means 1; MaxChoices must remain >= 2 for
// existing catalogs. Score requires an inclusive level
// range with minimum >= 2. Limits for unsupported kinds must be zero.
type ClassifierCapabilities struct {
	// Omitted locations retain context for existing Decisions image models.
	// Defaults are interpreted at call time, never materialized into catalogs.
	ImageLocations []ClassifierImageLocation `json:"imageLocations,omitempty"`
	MaxInputImages int                       `json:"maxInputImages,omitempty"`
	ImageMIMETypes []string                  `json:"imageMimeTypes,omitempty"`
	Kinds          []ClassifierQuestionKind  `json:"kinds"`
	MaxChoices     int                       `json:"maxChoices"`
	MinChoices     int                       `json:"minChoices,omitempty"`
	// Omitted formats allow the legacy string/object/array descriptions.
	DescriptionFormats []ClassifierDescriptionFormat `json:"descriptionFormats,omitempty"`
	MinScoreLevels     int                           `json:"minScoreLevels"`
	MaxScoreLevels     int                           `json:"maxScoreLevels"`
}

// ClassifierDescriptionFormat names a public JSON description's root type.
type ClassifierDescriptionFormat string

const (
	ClassifierDescriptionString ClassifierDescriptionFormat = "string"
	ClassifierDescriptionObject ClassifierDescriptionFormat = "object"
	ClassifierDescriptionArray  ClassifierDescriptionFormat = "array"
)

// ClassifierImageLocation names the visibility of an image use. Only context
// is implemented on Decisions and explicit experimental Jev candidates; other
// positions fail capability.
type ClassifierImageLocation string

const (
	ClassifierImageLocationContext      ClassifierImageLocation = "context"
	ClassifierImageLocationState        ClassifierImageLocation = "state"
	ClassifierImageLocationInstructions ClassifierImageLocation = "instructions"
	ClassifierImageLocationCriteria     ClassifierImageLocation = "criteria"
)
