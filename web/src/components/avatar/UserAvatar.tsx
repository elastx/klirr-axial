import { AvatarConfiguration } from "./types";
import { randomProperties, ColorSpace } from "./properties";
import { paths } from "./paths";
import { Avatar } from "@mantine/core";
import { defaultColorSpace } from "./defaultColorSpace";

type AvatarProps = {
  seed: string;
  colorSpace?: ColorSpace;
  size?: number;
};

type PathFunction = (color: string) => JSX.Element;
type ClothingFunction = (color: string, graphic: string) => JSX.Element;
type TopFunction = (color: string, hairColor: string) => JSX.Element;

type AvatarFunctions = {
  accessory: PathFunction;
  clothing: ClothingFunction;
  clothingGraphic: PathFunction;
  eyebrows: PathFunction;
  eyes: PathFunction;
  facialHair: PathFunction;
  mouth: PathFunction;
  nose: PathFunction;
  skin: PathFunction;
  top: TopFunction;
};

const avatarFunctions = (avatar: AvatarConfiguration): AvatarFunctions => {
  let accessory: PathFunction = paths.accessory.none;
  switch (avatar.accessory) {
    case "none":
      accessory = paths.accessory.none;
      break;
    case "kurt":
      accessory = paths.accessory.kurt;
      break;
    case "prescription01":
      accessory = paths.accessory.prescription01;
      break;
    case "prescription02":
      accessory = paths.accessory.prescription02;
      break;
    case "round":
      accessory = paths.accessory.round;
      break;
    case "sunglasses":
      accessory = paths.accessory.sunglasses;
      break;
    case "wayfarers":
      accessory = paths.accessory.wayfarers;
      break;
  }

  let clothing: ClothingFunction = paths.clothing.blazerAndShirt;
  switch (avatar.clothing) {
    case "blazerAndShirt":
      clothing = paths.clothing.blazerAndShirt;
      break;
    case "blazerAndSweater":
      clothing = paths.clothing.blazerAndSweater;
      break;
    case "collarAndSweater":
      clothing = paths.clothing.collarAndSweater;
      break;
    case "graphicShirt":
      clothing = paths.clothing.graphicShirt;
      break;
    case "hoodie":
      clothing = paths.clothing.hoodie;
      break;
    case "overall":
      clothing = paths.clothing.overall;
      break;
    case "shirtCrewNeck":
      clothing = paths.clothing.shirtCrewNeck;
      break;
    case "shirtScoopNeck":
      clothing = paths.clothing.shirtScoopNeck;
      break;
    case "shirtVNeck":
      clothing = paths.clothing.shirtVNeck;
      break;
  }

  let clothingGraphic: PathFunction = paths.clothingGraphic.bat;
  switch (avatar.clothingGraphic) {
    case "skrullOutline":
      clothingGraphic = paths.clothingGraphic.skrullOutline;
      break;
    case "skrull":
      clothingGraphic = paths.clothingGraphic.skrull;
      break;
    case "resist":
      clothingGraphic = paths.clothingGraphic.resist;
      break;
    case "pizza":
      clothingGraphic = paths.clothingGraphic.pizza;
      break;
    case "hola":
      clothingGraphic = paths.clothingGraphic.hola;
      break;
    case "diamond":
      clothingGraphic = paths.clothingGraphic.diamond;
      break;
    case "deer":
      clothingGraphic = paths.clothingGraphic.deer;
      break;
    case "dumbia":
      clothingGraphic = paths.clothingGraphic.dumbia;
      break;
    case "bear":
      clothingGraphic = paths.clothingGraphic.bear;
      break;
    case "bat":
      clothingGraphic = paths.clothingGraphic.bat;
      break;
  }

  let eyebrows: PathFunction = paths.eyebrows.default;
  switch (avatar.eyebrows) {
    case "angryNatural":
      eyebrows = paths.eyebrows.angryNatural;
      break;
    case "defaultNatural":
      eyebrows = paths.eyebrows.defaultNatural;
      break;
    case "flatNatural":
      eyebrows = paths.eyebrows.flatNatural;
      break;
    case "frownNatural":
      eyebrows = paths.eyebrows.frownNatural;
      break;
    case "raisedExcitedNatural":
      eyebrows = paths.eyebrows.raisedExcitedNatural;
      break;
    case "sadConcernedNatural":
      eyebrows = paths.eyebrows.sadConcernedNatural;
      break;
    case "unibrowNatural":
      eyebrows = paths.eyebrows.unibrowNatural;
      break;
    case "upDownNatural":
      eyebrows = paths.eyebrows.upDownNatural;
      break;
    case "raisedExcited":
      eyebrows = paths.eyebrows.raisedExcited;
      break;
    case "angry":
      eyebrows = paths.eyebrows.angry;
      break;
    case "default":
      eyebrows = paths.eyebrows.default;
      break;
    case "sadConcerned":
      eyebrows = paths.eyebrows.sadConcerned;
      break;
    case "upDown":
      eyebrows = paths.eyebrows.upDown;
      break;
  }

  let eyes: PathFunction = paths.eyes.default;
  switch (avatar.eyes) {
    case "squint":
      eyes = paths.eyes.squint;
      break;
    case "closed":
      eyes = paths.eyes.closed;
      break;
    case "cry":
      eyes = paths.eyes.cry;
      break;
    case "default":
      eyes = paths.eyes.default;
      break;
    case "eyeRoll":
      eyes = paths.eyes.eyeRoll;
      break;
    case "happy":
      eyes = paths.eyes.happy;
      break;
    case "hearts":
      eyes = paths.eyes.hearts;
      break;
    case "side":
      eyes = paths.eyes.side;
      break;
    case "surprised":
      eyes = paths.eyes.surprised;
      break;
    case "wink":
      eyes = paths.eyes.wink;
      break;
    case "winkWacky":
      eyes = paths.eyes.winkWacky;
      break;
    case "xDizzy":
      eyes = paths.eyes.xDizzy;
      break;
  }

  let facialHair: PathFunction = paths.facialHair.none;
  switch (avatar.facialHair) {
    case "none":
      facialHair = paths.facialHair.none;
      break;
    case "beardLight":
      facialHair = paths.facialHair.beardLight;
      break;
    case "beardMagestic":
      facialHair = paths.facialHair.beardMagestic;
      break;
    case "beardMedium":
      facialHair = paths.facialHair.beardMedium;
      break;
    case "moustaceFancy":
      facialHair = paths.facialHair.moustaceFancy;
      break;
    case "moustacheMagnum":
      facialHair = paths.facialHair.moustacheMagnum;
      break;
  }

  let mouth: PathFunction = paths.mouth.default;
  switch (avatar.mouth) {
    case "concerned":
      mouth = paths.mouth.concerned;
      break;
    case "default":
      mouth = paths.mouth.default;
      break;
    case "disbelief":
      mouth = paths.mouth.disbelief;
      break;
    case "eating":
      mouth = paths.mouth.eating;
      break;
    case "grimace":
      mouth = paths.mouth.grimace;
      break;
    case "sad":
      mouth = paths.mouth.sad;
      break;
    case "screamOpen":
      mouth = paths.mouth.screamOpen;
      break;
    case "serious":
      mouth = paths.mouth.serious;
      break;
    case "smile":
      mouth = paths.mouth.smile;
      break;
    case "tongue":
      mouth = paths.mouth.tongue;
      break;
    case "twinkle":
      mouth = paths.mouth.twinkle;
      break;
    case "vomit":
      mouth = paths.mouth.vomit;
      break;
  }

  let top: TopFunction = paths.top.theCaesar;
  switch (avatar.top) {
    case "eyepatch":
      top = paths.top.eyepatch;
      break;
    case "turban":
      top = paths.top.turban;
      break;
    case "hijab":
      top = paths.top.hijab;
      break;
    case "hat":
      top = paths.top.hat;
      break;
    case "winterHat01":
      top = paths.top.winterHat01;
      break;
    case "winterHat02":
      top = paths.top.winterHat02;
      break;
    case "winterHat03":
      top = paths.top.winterHat03;
      break;
    case "winterHat04":
      top = paths.top.winterHat04;
      break;
    case "bigHair":
      top = paths.top.bigHair;
      break;
    case "bob":
      top = paths.top.bob;
      break;
    case "bun":
      top = paths.top.bun;
      break;
    case "curly":
      top = paths.top.curly;
      break;
    case "curvy":
      top = paths.top.curvy;
      break;
    case "dreads":
      top = paths.top.dreads;
      break;
    case "frida":
      top = paths.top.frida;
      break;
    case "froAndBand":
      top = paths.top.froAndBand;
      break;
    case "fro":
      top = paths.top.fro;
      break;
    case "longButNotTooLong":
      top = paths.top.longButNotTooLong;
      break;
    case "miaWallace":
      top = paths.top.miaWallace;
      break;
    case "shavedSides":
      top = paths.top.shavedSides;
      break;
    case "straightAndStrand":
      top = paths.top.straightAndStrand;
      break;
    case "straight01":
      top = paths.top.straight01;
      break;
    case "straight02":
      top = paths.top.straight02;
      break;
    case "dreads01":
      top = paths.top.dreads01;
      break;
    case "dreads02":
      top = paths.top.dreads02;
      break;
    case "frizzle":
      top = paths.top.frizzle;
      break;
    case "shaggyMullet":
      top = paths.top.shaggyMullet;
      break;
    case "shaggy":
      top = paths.top.shaggy;
      break;
    case "shortCurly":
      top = paths.top.shortCurly;
      break;
    case "shortFlat":
      top = paths.top.shortFlat;
      break;
    case "shortRound":
      top = paths.top.shortRound;
      break;
    case "sides":
      top = paths.top.sides;
      break;
    case "shortWaved":
      top = paths.top.shortWaved;
      break;
    case "theCaesarAndSidePart":
      top = paths.top.theCaesarAndSidePart;
      break;
    case "theCaesar":
      top = paths.top.theCaesar;
      break;
  }

  const nose: PathFunction = paths.nose.default;
  const skin: PathFunction = paths.skin.default;

  return {
    accessory,
    clothing,
    clothingGraphic,
    eyebrows,
    eyes,
    facialHair,
    mouth,
    nose,
    skin,
    top,
  };
};
const getDarkestColor = (colors: string[]): string => {
  let darkest = 300;
  let darkestIndex = 0;
  colors.forEach((color, index) => {
    const match = color.match(/\d+/g);
    if (!match) return;
    const [_, __, l] = match;
    if (Number(l) < darkest) {
      darkest = Number(l);
      darkestIndex = index;
    }
  });
  return colors[darkestIndex];
};

const getLightestColor = (colors: string[]): string => {
  let lightest = 0;
  let lightestIndex = 0;
  colors.forEach((color, index) => {
    const match = color.match(/\d+/g);
    if (!match) return;
    const [_, __, l] = match;
    if (Number(l) > lightest) {
      lightest = Number(l);
      lightestIndex = index;
    }
  });
  return colors[lightestIndex];
};

const UserAvatar: React.FC<AvatarProps> = ({
  seed,
  colorSpace = defaultColorSpace,
  size = 80,
}) => {
  const avatar = randomProperties(seed, colorSpace);

  const group = (content: JSX.Element, x: number, y: number) => {
    return <g transform={`translate(${x}, ${y})`}>{content}</g>;
  };

  const funcs = avatarFunctions(avatar);

  const positions: Record<string, { x: number; y: number }> = {
    skin: { x: 40, y: 36 },
    clothing: { x: 8, y: 170 },
    mouth: { x: 86, y: 134 },
    nose: { x: 112, y: 122 },
    eyes: { x: 84, y: 90 },
    eyebrows: { x: 84, y: 82 },
    top: { x: 7, y: 0 },
    facialHair: { x: 56, y: 72 },
    accessories: { x: 69, y: 85 },
  };

  const darkestColor = getDarkestColor([
    avatar.skinColor,
    avatar.clothingColor,
    avatar.facialHairColor,
    avatar.topColor,
    avatar.accessoryColor,
  ]);

  const lightestColor = getLightestColor([
    avatar.skinColor,
    avatar.clothingColor,
    avatar.facialHairColor,
    avatar.topColor,
    avatar.accessoryColor,
  ]);

  return (
    <Avatar
      size={size}
      gradient={{ from: darkestColor, to: lightestColor }}
      variant="gradient"
    >
      <svg
        viewBox="0 0 280 280"
        xmlns="http://www.w3.org/2000/svg"
        style={{
          width: "100%",
          height: "100%",
          filter: "drop-shadow(rgba(0, 0, 0, 0.6) 3px 10px 12px)",
        }}
      >
        {group(
          funcs.skin(avatar.skinColor),
          positions.skin.x,
          positions.skin.y
        )}
        {group(
          funcs.clothing(avatar.clothingColor, avatar.clothingGraphic),
          positions.clothing.x,
          positions.clothing.y
        )}
        {group(
          funcs.mouth(avatar.skinColor),
          positions.mouth.x,
          positions.mouth.y
        )}
        {group(
          funcs.nose(avatar.skinColor),
          positions.nose.x,
          positions.nose.y
        )}
        {group(
          funcs.eyes(avatar.skinColor),
          positions.eyes.x,
          positions.eyes.y
        )}
        {group(
          funcs.eyebrows(avatar.hairColor),
          positions.eyebrows.x,
          positions.eyebrows.y
        )}
        {group(
          funcs.top(avatar.topColor, avatar.hairColor),
          positions.top.x,
          positions.top.y
        )}
        {group(
          funcs.facialHair(avatar.facialHairColor),
          positions.facialHair.x,
          positions.facialHair.y
        )}
        {group(
          funcs.accessory(avatar.accessoryColor),
          positions.accessories.x,
          positions.accessories.y
        )}
      </svg>
    </Avatar>
  );
};

export default UserAvatar;
