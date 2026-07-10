package appwindow

// Material is an NSVisualEffectMaterial: the tint AppKit blends over the
// blurred desktop behind a translucent window. The blur radius is fixed.
type Material int

// MaterialUnderWindowBackground tints the least of the available materials.
const MaterialUnderWindowBackground Material = 21
