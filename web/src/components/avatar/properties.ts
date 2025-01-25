import { Avatar } from "./types";

const randomHSLFromString = (seed: string): string => {
  // Create a simple hash from the string
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = (hash << 5) - hash + seed.charCodeAt(i);
    hash = hash & hash; // Convert to 32-bit integer
  }

  // Use the hash to generate HSL values within pleasing ranges
  const hue = Math.abs(hash) % 360; // 0-360 degrees
  const saturation = 35 + (Math.abs(hash >> 8) % 50); // 35-85%
  const lightness = 45 + (Math.abs(hash >> 16) % 20); // 45-65%

  return `hsl(${hue}, ${saturation}%, ${lightness}%)`;
};

/**
 * Selects a random value from a string union type using a seed
 * @param seed String used to generate deterministic random selection
 * @returns A random value from the string union type T
 */
function randomFromStringUnion<T extends string>(seed: string): T {
  // Get array of string literal types from T
  const values = Object.values(
    // This trick forces TypeScript to preserve the string literal types
    {} as { [K in T]: K }
  );

  // Use seed to generate deterministic random index
  const index = Math.floor(seededRandom(seed) * values.length);

  return values[index] as T;
}

/**
 * Generate a deterministic random number between 0 and 1 using a string seed
 */
function seededRandom(seed: string): number {
  // Create a numeric hash from the string
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = (hash << 5) - hash + seed.charCodeAt(i);
    hash = hash & hash; // Convert to 32-bit integer
  }

  const x = Math.sin(hash) * 10000;
  return x - Math.floor(x);
}

export const randomProperties = (seed: string): Avatar => {
  return {
    accessory: randomFromStringUnion<Avatar["accessory"]>(seed + "-accessory"),
    clothing: randomFromStringUnion<Avatar["clothing"]>(seed + "-clothing"),
    clothingGraphic: randomFromStringUnion<Avatar["clothingGraphic"]>(
      seed + "-clothingGraphic"
    ),
    eyebrows: randomFromStringUnion<Avatar["eyebrows"]>(seed + "-eyebrows"),
    eyes: randomFromStringUnion<Avatar["eyes"]>(seed + "-eyes"),
    facialHair: randomFromStringUnion<Avatar["facialHair"]>(
      seed + "-facialHair"
    ),
    mouth: randomFromStringUnion<Avatar["mouth"]>(seed + "-mouth"),
    nose: randomFromStringUnion<Avatar["nose"]>(seed + "-nose"),
    skin: randomFromStringUnion<Avatar["skin"]>(seed + "-skin"),
    top: randomFromStringUnion<Avatar["top"]>(seed + "-top"),
    skinColor: randomHSLFromString(seed + "-skinColor"),
    hairColor: randomHSLFromString(seed + "-hairColor"),
    facialHairColor: randomHSLFromString(seed + "-facialHairColor"),
    topColor: randomHSLFromString(seed + "-topColor"),
    clothingColor: randomHSLFromString(seed + "-clothingColor"),
    graphicColor: randomHSLFromString(seed + "-graphicColor"),
    accessoryColor: randomHSLFromString(seed + "-accessoryColor"),
  };
};
