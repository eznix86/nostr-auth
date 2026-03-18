import images from "../../images/images.json";

const BACKGROUND_VARIANTS = import.meta.glob("../../images/*.webp", {
  eager: true,
  import: "default",
});

export default function Background({ branding }) {
  const defaultVariant = images.default || "canyon-falls";
  const variant = branding?.background?.source?.type === "preset" ? branding.background.source.variant : defaultVariant;
  const backgroundUrl = resolveBackgroundUrl(variant) || resolveBackgroundUrl(defaultVariant);

  return (
    <div
      className="absolute -inset-8 bg-cover bg-center bg-no-repeat"
      style={{
        backgroundImage: `linear-gradient(to bottom, rgba(255, 255, 255, 0.02), rgba(7, 10, 8, 0.08)), url(${backgroundUrl})`,
        maskImage: "linear-gradient(to bottom, rgba(0, 0, 0, 0.98) 0%, rgba(0, 0, 0, 0.95) 48%, rgba(0, 0, 0, 0.82) 74%, rgba(0, 0, 0, 0.5) 92%, transparent 100%)",
        WebkitMaskImage:
          "linear-gradient(to bottom, rgba(0, 0, 0, 0.98) 0%, rgba(0, 0, 0, 0.95) 48%, rgba(0, 0, 0, 0.82) 74%, rgba(0, 0, 0, 0.5) 92%, transparent 100%)",
      }}
    />
  );
}

function resolveBackgroundUrl(variant) {
  const file = images.variants?.[variant]?.file;
  if (!file) {
    return null;
  }

  return BACKGROUND_VARIANTS[`../../images/${file}`] || null;
}
