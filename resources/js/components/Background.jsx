import canyonFalls from "../../images/canyon-falls.jpg";
import fieldsRoad from "../../images/fields-road.jpg";
import mountainValley from "../../images/mountain-valley.jpg";
import stormValley from "../../images/storm-valley.jpg";

const DEFAULT_VARIANT = "canyon-falls";

const BACKGROUND_VARIANTS = {
  "canyon-falls": canyonFalls,
  "fields-road": fieldsRoad,
  "mountain-valley": mountainValley,
  "storm-valley": stormValley,
};

export default function Background({ branding }) {
  const variant = branding?.background?.source?.type === "preset" ? branding.background.source.variant : DEFAULT_VARIANT;
  const backgroundUrl = BACKGROUND_VARIANTS[variant] || BACKGROUND_VARIANTS[DEFAULT_VARIANT];

  return (
    <div
      className="absolute inset-[-2rem] bg-cover bg-center bg-no-repeat"
      style={{
        backgroundImage: `linear-gradient(to bottom, rgba(255, 255, 255, 0.02), rgba(7, 10, 8, 0.08)), url(${backgroundUrl})`,
        maskImage: "linear-gradient(to bottom, rgba(0, 0, 0, 0.98) 0%, rgba(0, 0, 0, 0.95) 48%, rgba(0, 0, 0, 0.82) 74%, rgba(0, 0, 0, 0.5) 92%, transparent 100%)",
        WebkitMaskImage:
          "linear-gradient(to bottom, rgba(0, 0, 0, 0.98) 0%, rgba(0, 0, 0, 0.95) 48%, rgba(0, 0, 0, 0.82) 74%, rgba(0, 0, 0, 0.5) 92%, transparent 100%)",
      }}
    />
  );
}
