export type Avatar = {
  accessory: Accessory;
  clothing: Clothing;
  clothingGraphic: ClothingGraphic;
  eyebrows: Eyebrows;
  eyes: Eyes;
  facialHair: FacialHair;
  mouth: Mouth;
  nose: Nose;
  skin: Skin;
  top: Top;
  skinColor: string;
  hairColor: string;
  facialHairColor: string;
  topColor: string;
  clothingColor: string;
  graphicColor: string;
  accessoryColor: string;
};

export enum Accessory {
  none = "none",
  kurt = "kurt",
  prescription01 = "prescription01",
  prescription02 = "prescription02",
  round = "round",
  sunglasses = "sunglasses",
  wayfarers = "wayfarers",
}
export enum Clothing {
  blazerAndShirt = "blazerAndShirt",
  blazerAndSweater = "blazerAndSweater",
  collarAndSweater = "collarAndSweater",
  graphicShirt = "graphicShirt",
  hoodie = "hoodie",
  overall = "overall",
  shirtCrewNeck = "shirtCrewNeck",
  shirtScoopNeck = "shirtScoopNeck",
  shirtVNeck = "shirtVNeck",
}
export enum ClothingGraphic {
  skrullOutline = "skrullOutline",
  skrull = "skrull",
  resist = "resist",
  pizza = "pizza",
  hola = "hola",
  diamond = "diamond",
  deer = "deer",
  dumbia = "dumbia",
  bear = "bear",
  bat = "bat",
}
export enum Eyebrows {
  angryNatural = "angryNatural",
  defaultNatural = "defaultNatural",
  flatNatural = "flatNatural",
  frownNatural = "frownNatural",
  raisedExcitedNatural = "raisedExcitedNatural",
  sadConcernedNatural = "sadConcernedNatural",
  unibrowNatural = "unibrowNatural",
  upDownNatural = "upDownNatural",
  raisedExcited = "raisedExcited",
  angry = "angry",
  default = "default",
  sadConcerned = "sadConcerned",
  upDown = "upDown",
}
export enum Eyes {
  squint = "squint",
  closed = "closed",
  cry = "cry",
  default = "default",
  eyeRoll = "eyeRoll",
  happy = "happy",
  hearts = "hearts",
  side = "side",
  surprised = "surprised",
  wink = "wink",
  winkWacky = "winkWacky",
  xDizzy = "xDizzy",
}
export enum FacialHair {
  none = "none",
  beardLight = "beardLight",
  beardMagestic = "beardMagestic",
  beardMedium = "beardMedium",
  moustaceFancy = "moustaceFancy",
  moustacheMagnum = "moustacheMagnum",
}
export enum Mouth {
  concerned = "concerned",
  default = "default",
  disbelief = "disbelief",
  eating = "eating",
  grimace = "grimace",
  sad = "sad",
  screamOpen = "screamOpen",
  serious = "serious",
  smile = "smile",
  tongue = "tongue",
  twinkle = "twinkle",
  vomit = "vomit",
}
export enum Nose {
  default = "default",
}
export enum Skin {
  default = "default",
}
export enum Top {
  eyepatch = "eyepatch",
  turban = "turban",
  hijab = "hijab",
  hat = "hat",
  winterHat01 = "winterHat01",
  winterHat02 = "winterHat02",
  winterHat03 = "winterHat03",
  winterHat04 = "winterHat04",
  bigHair = "bigHair",
  bob = "bob",
  bun = "bun",
  curly = "curly",
  curvy = "curvy",
  dreads = "dreads",
  frida = "frida",
  froAndBand = "froAndBand",
  fro = "fro",
  longButNotTooLong = "longButNotTooLong",
  miaWallace = "miaWallace",
  shavedSides = "shavedSides",
  straightAndStrand = "straightAndStrand",
  straight01 = "straight01",
  straight02 = "straight02",
  dreads01 = "dreads01",
  dreads02 = "dreads02",
  frizzle = "frizzle",
  shaggyMullet = "shaggyMullet",
  shaggy = "shaggy",
  shortCurly = "shortCurly",
  shortFlat = "shortFlat",
  shortRound = "shortRound",
  sides = "sides",
  shortWaved = "shortWaved",
  theCaesarAndSidePart = "theCaesarAndSidePart",
  theCaesar = "theCaesar",
}
