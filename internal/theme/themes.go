package theme

const DefaultThemeName = "catppuccin-mocha"

// AllThemes is the ordered list shown when cycling themes.
var AllThemes = []Theme{
	CatppuccinMocha,
	CatppuccinLatte,
	Dracula,
	Nord,
	GruvboxDark,
	GruvboxLight,
	TokyoNight,
	RosePine,
}

var CatppuccinMocha = Theme{
	Name: "catppuccin-mocha",
	Base: BaseColors{
		Text:       "#cdd6f4",
		Subtext:    "#bac2de",
		Muted:      "#9399b2",
		Surface:    "#45475a",
		SurfaceAlt: "#585b70",
		Primary:    "#b4befe",
		Red:        "#f38ba8",
		Orange:     "#fab387",
		Yellow:     "#f9e2af",
		Green:      "#a6e3a1",
		Blue:       "#89b4fa",
		Purple:     "#cba6f7",
		Pink:       "#f5c2e7",
	},
}

var CatppuccinLatte = Theme{
	Name: "catppuccin-latte",
	Base: BaseColors{
		Text:       "#4c4f69",
		Subtext:    "#6c6f85",
		Muted:      "#9ca0b0",
		Surface:    "#ccd0da",
		SurfaceAlt: "#bcc0cc",
		Primary:    "#7287fd",
		Red:        "#d20f39",
		Orange:     "#fe640b",
		Yellow:     "#df8e1d",
		Green:      "#40a02b",
		Blue:       "#1e66f5",
		Purple:     "#8839ef",
		Pink:       "#ea76cb",
	},
}

var Dracula = Theme{
	Name: "dracula",
	Base: BaseColors{
		Text:       "#f8f8f2",
		Subtext:    "#e2e2dc",
		Muted:      "#6272a4",
		Surface:    "#44475a",
		SurfaceAlt: "#6272a4",
		Primary:    "#bd93f9",
		Red:        "#ff5555",
		Orange:     "#ffb86c",
		Yellow:     "#f1fa8c",
		Green:      "#50fa7b",
		Blue:       "#8be9fd",
		Purple:     "#bd93f9",
		Pink:       "#ff79c6",
	},
}

var Nord = Theme{
	Name: "nord",
	Base: BaseColors{
		Text:       "#eceff4",
		Subtext:    "#e5e9f0",
		Muted:      "#4c566a",
		Surface:    "#3b4252",
		SurfaceAlt: "#434c5e",
		Primary:    "#88c0d0",
		Red:        "#bf616a",
		Orange:     "#d08770",
		Yellow:     "#ebcb8b",
		Green:      "#a3be8c",
		Blue:       "#81a1c1",
		Purple:     "#b48ead",
		Pink:       "#b48ead",
	},
}

var GruvboxDark = Theme{
	Name: "gruvbox-dark",
	Base: BaseColors{
		Text:       "#ebdbb2",
		Subtext:    "#d5c4a1",
		Muted:      "#a89984",
		Surface:    "#3c3836",
		SurfaceAlt: "#504945",
		Primary:    "#83a598",
		Red:        "#fb4934",
		Orange:     "#fe8019",
		Yellow:     "#fabd2f",
		Green:      "#b8bb26",
		Blue:       "#83a598",
		Purple:     "#d3869b",
		Pink:       "#d3869b",
	},
}

var GruvboxLight = Theme{
	Name: "gruvbox-light",
	Base: BaseColors{
		Text:       "#3c3836",
		Subtext:    "#504945",
		Muted:      "#7c6f64",
		Surface:    "#d5c4a1",
		SurfaceAlt: "#bdae93",
		Primary:    "#076678",
		Red:        "#9d0006",
		Orange:     "#af3a03",
		Yellow:     "#b57614",
		Green:      "#79740e",
		Blue:       "#076678",
		Purple:     "#8f3f71",
		Pink:       "#8f3f71",
	},
}

var TokyoNight = Theme{
	Name: "tokyo-night",
	Base: BaseColors{
		Text:       "#c0caf5",
		Subtext:    "#a9b1d6",
		Muted:      "#565f89",
		Surface:    "#292e42",
		SurfaceAlt: "#3b4261",
		Primary:    "#7aa2f7",
		Red:        "#f7768e",
		Orange:     "#ff9e64",
		Yellow:     "#e0af68",
		Green:      "#9ece6a",
		Blue:       "#7aa2f7",
		Purple:     "#bb9af7",
		Pink:       "#ff007c",
	},
}

var RosePine = Theme{
	Name: "rose-pine",
	Base: BaseColors{
		Text:       "#e0def4",
		Subtext:    "#908caa",
		Muted:      "#6e6a86",
		Surface:    "#26233a",
		SurfaceAlt: "#1f1d2e",
		Primary:    "#c4a7e7",
		Red:        "#eb6f92",
		Orange:     "#f6c177",
		Yellow:     "#f6c177",
		Green:      "#31748f",
		Blue:       "#9ccfd8",
		Purple:     "#c4a7e7",
		Pink:       "#ebbcba",
	},
}
