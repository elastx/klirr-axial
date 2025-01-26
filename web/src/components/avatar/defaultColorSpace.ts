import { ColorSpace } from "./properties";

export const defaultColorSpace: ColorSpace = {
  skin: {
    hueMin: 0,
    hueMax: 40,
    saturationMin: 20,
    saturationMax: 60,
    lightnessMin: 6,
    lightnessMax: 80,
  },
  hair: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 0,
    saturationMax: 100,
    lightnessMin: 0,
    lightnessMax: 45,
  },
  facialHair: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 30,
    saturationMax: 100,
    lightnessMin: 0,
    lightnessMax: 45,
  },
  clothing: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 58,
    saturationMax: 100,
    lightnessMin: 23,
    lightnessMax: 80,
  },
  accessory: {
    hueMin: 0,
    hueMax: 360,
    saturationMin: 56,
    saturationMax: 100,
    lightnessMin: 28,
    lightnessMax: 54,
  },
};
