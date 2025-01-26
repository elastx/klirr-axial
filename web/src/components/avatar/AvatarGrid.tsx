import React, { useState, useMemo, useCallback } from "react";
import UserAvatar from "./UserAvatar";
import { defaultColorSpace } from "./defaultColorSpace";
import { ColorSpace } from "./properties";
import { Group, RangeSlider, Container, Textarea, Box } from "@mantine/core";

type AvatarGridProps = {
  rows?: number;
  cols?: number;
  avatarSize?: number;
};

type ColorTypeControls = {
  label: string;
  key: keyof ColorSpace;
};

const colorTypes: ColorTypeControls[] = [
  { label: "Skin", key: "skin" },
  { label: "Hair", key: "hair" },
  { label: "Facial Hair", key: "facialHair" },
  { label: "Clothing", key: "clothing" },
  { label: "Accessory", key: "accessory" },
];

const AvatarGrid: React.FC<AvatarGridProps> = ({
  rows = 10,
  cols = 8,
  avatarSize = 100,
}) => {
  const [colorSpace, setColorSpace] = useState<ColorSpace>(defaultColorSpace);

  const handleRangeChange = useCallback(
    (
      colorType: keyof ColorSpace,
      property: "hue" | "saturation" | "lightness",
      values: [number, number]
    ) => {
      setColorSpace((prev) => ({
        ...prev,
        [colorType]: {
          ...prev[colorType],
          [`${property}Min`]: values[0],
          [`${property}Max`]: values[1],
        },
      }));
    },
    []
  );

  const generateSeeds = () => {
    const seeds = [];
    for (let i = 0; i < rows * cols; i++) {
      seeds.push(`avatar-${i}`);
    }
    return seeds;
  };

  const getHueGradient = useMemo(() => {
    const stops = [];
    for (let i = 0; i <= 360; i += 60) {
      stops.push(`hsl(${i}, 100%, 50%)`);
    }
    return `linear-gradient(to right, ${stops.join(", ")})`;
  }, []);

  const getSaturationGradient = useCallback((hue: number) => {
    return `linear-gradient(to right, hsl(${hue}, 0%, 50%), hsl(${hue}, 100%, 50%))`;
  }, []);

  const getLightnessGradient = useCallback((hue: number) => {
    return `linear-gradient(to right, hsl(${hue}, 100%, 0%), hsl(${hue}, 100%, 50%), hsl(${hue}, 100%, 100%))`;
  }, []);

  const ColorControls = React.memo(({ type }: { type: ColorTypeControls }) => {
    const midHue = Math.floor(
      (colorSpace[type.key].hueMin + colorSpace[type.key].hueMax) / 2
    );

    return (
      <div className="p-4 bg-gray-50 rounded" style={{ width: 300 }}>
        <h3 className="font-bold mb-4">{type.label}</h3>
        <div className="space-y-6">
          <div>
            <div className="text-sm mb-2">Hue (0-360)</div>
            <div
              className="h-4 rounded mb-2"
              style={{
                background: getHueGradient,
                height: "10px",
                width: "100%",
              }}
            />
            <RangeSlider
              value={[colorSpace[type.key].hueMin, colorSpace[type.key].hueMax]}
              onChange={(values) => handleRangeChange(type.key, "hue", values)}
              min={0}
              max={360}
              label={(value) => value}
            />
          </div>
          <div>
            <div className="text-sm mb-2">Saturation (0-100)</div>
            <div
              className="h-4 rounded mb-2"
              style={{
                background: getSaturationGradient(midHue),
                height: "10px",
                width: "100%",
              }}
            />
            <RangeSlider
              value={[
                colorSpace[type.key].saturationMin,
                colorSpace[type.key].saturationMax,
              ]}
              onChange={(values) =>
                handleRangeChange(type.key, "saturation", values)
              }
              min={0}
              max={100}
              label={(value) => `${value}%`}
            />
          </div>
          <div>
            <div className="text-sm mb-2">Lightness (0-100)</div>
            <div
              className="h-4 rounded mb-2"
              style={{
                background: getLightnessGradient(midHue),
                height: "10px",
                width: "100%",
              }}
            />
            <RangeSlider
              value={[
                colorSpace[type.key].lightnessMin,
                colorSpace[type.key].lightnessMax,
              ]}
              onChange={(values) =>
                handleRangeChange(type.key, "lightness", values)
              }
              min={0}
              max={100}
              label={(value) => `${value}%`}
            />
          </div>
        </div>
      </div>
    );
  });

  const colorSpaceCode = useMemo(() => {
    return `export const defaultColorSpace: ColorSpace = ${JSON.stringify(
      colorSpace,
      null,
      2
    )};`;
  }, [colorSpace]);

  return (
    <Container
      size="100%"
      p={0}
      style={{ overflow: "auto", height: "calc(100vh - 100px)" }}
    >
      <div className="space-y-8">
        <Group align="flex-start" justify="center" gap="md">
          {colorTypes.map((type) => (
            <ColorControls key={type.key} type={type} />
          ))}
        </Group>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: `repeat(${cols}, ${avatarSize}px)`,
            gap: "8px",
            justifyContent: "center",
          }}
        >
          {generateSeeds().map((seed) => (
            <UserAvatar
              key={seed}
              seed={seed}
              colorSpace={colorSpace}
              size={avatarSize}
            />
          ))}
        </div>

        <Box mx="auto">
          <Textarea
            label="Generated Color Space Configuration"
            value={colorSpaceCode}
            minRows={15}
            autosize
            styles={{
              input: {
                fontFamily: "monospace",
                fontSize: "0.9em",
              },
            }}
          />
        </Box>
      </div>
    </Container>
  );
};

export default AvatarGrid;
