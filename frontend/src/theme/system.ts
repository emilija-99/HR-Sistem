import { createSystem, defaultConfig, defineConfig } from "@chakra-ui/react";

/**
 * Brend paleta (dizajn specifikacija):
 *
 * | Upotreba                | Boja          | HEX      | Token        |
 * | ----------------------- | ------------- | -------- | ------------ |
 * | Tamni navy (primarni)   | Dark navy     | #123B68  | `brand.800`  |
 * | Primarna plava          | Primary blue  | #1674D1  | `brand.600`  |
 * | Akcenat (svetlija plava)| Bright blue   | #2589E8  | `brand.500`  |
 * | Sekundarni tekst        | Text gray-blue| #5D7190  | `fg.muted`   |
 * | Pozadina (kartice)      | White         | #FFFFFF  | `bg.panel`   |
 * | Pozadina aplikacije     | Soft blue     | #F4F8FC  | `bg.subtle`  |
 *
 * `brand.500/600/700` su izabrani tako da Chakra semantika „legne" na dizajn:
 * `colorPalette="brand".solid` = `brand.600` (#1674D1), `.focusRing` = `brand.500`,
 * `.fg` = `brand.700`, `.subtle` = `brand.100`.
 */
export const BRAND = {
  navy: "#123B68",
  primary: "#1674D1",
  accent: "#2589E8",
  textMuted: "#5D7190",
  white: "#FFFFFF",
  soft: "#F4F8FC",
} as const;

const fontStack =
  'Inter, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif';

const config = defineConfig({
  theme: {
    tokens: {
      colors: {
        brand: {
          50: { value: "#F4F8FC" }, // soft blue — pozadina
          100: { value: "#E4EEFA" },
          200: { value: "#C7DCF4" },
          300: { value: "#9DC3EC" },
          400: { value: "#4A9AEA" },
          500: { value: "#2589E8" }, // akcenat
          600: { value: "#1674D1" }, // primarna
          700: { value: "#125FAE" }, // hover (tamnija od primarne)
          800: { value: "#123B68" }, // navy
          900: { value: "#0C2947" },
        },
      },
      fonts: {
        heading: { value: fontStack },
        body: { value: fontStack },
      },
    },
    semanticTokens: {
      colors: {
        // colorPalette="brand" (vidi layer-styles: solid/subtle/fg/...)
        brand: {
          contrast: { value: { _light: "#FFFFFF", _dark: "#FFFFFF" } },
          fg: {
            value: { _light: "{colors.brand.700}", _dark: "{colors.brand.300}" },
          },
          subtle: {
            value: { _light: "{colors.brand.100}", _dark: "{colors.brand.900}" },
          },
          muted: {
            value: { _light: "{colors.brand.200}", _dark: "{colors.brand.800}" },
          },
          emphasized: {
            value: { _light: "{colors.brand.300}", _dark: "{colors.brand.700}" },
          },
          solid: {
            value: { _light: "{colors.brand.600}", _dark: "{colors.brand.500}" },
          },
          focusRing: {
            value: { _light: "{colors.brand.500}", _dark: "{colors.brand.500}" },
          },
          border: {
            value: { _light: "{colors.brand.500}", _dark: "{colors.brand.400}" },
          },
        },
        // Tekst: navy kao primarni, gray-blue za sekundarne informacije
        fg: {
          DEFAULT: {
            value: { _light: "{colors.brand.800}", _dark: "{colors.gray.50}" },
          },
          muted: {
            value: { _light: "#5D7190", _dark: "{colors.gray.400}" },
          },
        },
        // Pozadina aplikacije je soft blue; kartice ostaju `bg.panel` (bela)
        bg: {
          subtle: {
            value: { _light: "{colors.brand.50}", _dark: "{colors.gray.950}" },
          },
          muted: {
            value: { _light: "{colors.brand.100}", _dark: "{colors.gray.900}" },
          },
        },
        // Ivicе u plavom tonu umesto neutralno sivih
        border: {
          DEFAULT: {
            value: { _light: "{colors.brand.200}", _dark: "{colors.brand.900}" },
          },
          muted: {
            value: { _light: "{colors.brand.100}", _dark: "{colors.brand.900}" },
          },
        },
      },
    },
  },
  globalCss: {
    "html, body": {
      fontFamily: "body",
      bg: "bg.subtle",
      color: "fg",
    },
  },
});

export const system = createSystem(defaultConfig, config);
