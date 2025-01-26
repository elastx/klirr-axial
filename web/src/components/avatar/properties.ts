import {
  AvatarConfiguration,
  Accessory,
  Clothing,
  ClothingGraphic,
  Eyebrows,
  Eyes,
  FacialHair,
  Mouth,
  Nose,
  Skin,
  Top,
} from "./types";

export type ColorSpace = {
  skin: {
    hueMin: number;
    hueMax: number;
    saturationMin: number;
    saturationMax: number;
    lightnessMin: number;
    lightnessMax: number;
  };
  hair: {
    hueMin: number;
    hueMax: number;
    saturationMin: number;
    saturationMax: number;
    lightnessMin: number;
    lightnessMax: number;
  };
  facialHair: {
    hueMin: number;
    hueMax: number;
    saturationMin: number;
    saturationMax: number;
    lightnessMin: number;
    lightnessMax: number;
  };
  clothing: {
    hueMin: number;
    hueMax: number;
    saturationMin: number;
    saturationMax: number;
    lightnessMin: number;
    lightnessMax: number;
  };
  accessory: {
    hueMin: number;
    hueMax: number;
    saturationMin: number;
    saturationMax: number;
    lightnessMin: number;
    lightnessMax: number;
  };
};

const defaultColorSpace: ColorSpace = {
  skin: {
    hueMin: 0,
    hueMax: 40,
    saturationMin: 20,
    saturationMax: 60,
    lightnessMin: 50,
    lightnessMax: 80,
  },
  hair: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 35,
    saturationMax: 85,
    lightnessMin: 15,
    lightnessMax: 45,
  },
  facialHair: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 35,
    saturationMax: 85,
    lightnessMin: 15,
    lightnessMax: 45,
  },
  clothing: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 35,
    saturationMax: 85,
    lightnessMin: 35,
    lightnessMax: 65,
  },
  accessory: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 35,
    saturationMax: 85,
    lightnessMin: 35,
    lightnessMax: 65,
  },
};

const randomHSLFromString = (
  seed: string,
  colorType: keyof ColorSpace,
  colorSpace: ColorSpace = defaultColorSpace
): string => {
  // Get three random values between 0-1 using the seed
  const hueRand = seededRandom(seed + "-hue");
  const satRand = seededRandom(seed + "-sat");
  const lightRand = seededRandom(seed + "-light");

  const space = colorSpace[colorType];

  // Map the random values between min and max ranges
  const hue = space.hueMin + hueRand * (space.hueMax - space.hueMin);
  const saturation =
    space.saturationMin + satRand * (space.saturationMax - space.saturationMin);
  const lightness =
    space.lightnessMin + lightRand * (space.lightnessMax - space.lightnessMin);

  return `hsl(${Math.round(hue)}, ${Math.round(saturation)}%, ${Math.round(
    lightness
  )}%)`;
};

/**
 * Selects a random value from a string union type using a seed
 * @param seed String used to generate deterministic random selection
 * @returns A random value from the string union type T
 */
function randomFromStringUnion<T extends string>(
  seed: string,
  enumType: { [key: string]: string }
): T {
  // Get array of enum values
  const values = Object.values(enumType);

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

export const randomProperties = (
  seed: string,
  colorSpace?: ColorSpace
): AvatarConfiguration => {
  return {
    accessory: randomFromStringUnion<AvatarConfiguration["accessory"]>(
      seed + "-accessory",
      Accessory
    ),
    clothing: randomFromStringUnion<AvatarConfiguration["clothing"]>(
      seed + "-clothing",
      Clothing
    ),
    clothingGraphic: randomFromStringUnion<
      AvatarConfiguration["clothingGraphic"]
    >(seed + "-clothingGraphic", ClothingGraphic),
    eyebrows: randomFromStringUnion<AvatarConfiguration["eyebrows"]>(
      seed + "-eyebrows",
      Eyebrows
    ),
    eyes: randomFromStringUnion<AvatarConfiguration["eyes"]>(
      seed + "-eyes",
      Eyes
    ),
    facialHair: randomFromStringUnion<AvatarConfiguration["facialHair"]>(
      seed + "-facialHair",
      FacialHair
    ),
    mouth: randomFromStringUnion<AvatarConfiguration["mouth"]>(
      seed + "-mouth",
      Mouth
    ),
    nose: randomFromStringUnion<AvatarConfiguration["nose"]>(
      seed + "-nose",
      Nose
    ),
    skin: randomFromStringUnion<AvatarConfiguration["skin"]>(
      seed + "-skin",
      Skin
    ),
    top: randomFromStringUnion<AvatarConfiguration["top"]>(seed + "-top", Top),
    skinColor: randomHSLFromString(seed + "-skinColor", "skin", colorSpace),
    hairColor: randomHSLFromString(seed + "-hairColor", "hair", colorSpace),
    facialHairColor: randomHSLFromString(
      seed + "-facialHairColor",
      "facialHair",
      colorSpace
    ),
    topColor: randomHSLFromString(seed + "-topColor", "hair", colorSpace),
    clothingColor: randomHSLFromString(
      seed + "-clothingColor",
      "clothing",
      colorSpace
    ),
    graphicColor: randomHSLFromString(
      seed + "-graphicColor",
      "clothing",
      colorSpace
    ),
    accessoryColor: randomHSLFromString(
      seed + "-accessoryColor",
      "accessory",
      colorSpace
    ),
  };
};
