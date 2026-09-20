import { HStack, Text } from "@chakra-ui/react";
import type { SystemStyleObject } from "@chakra-ui/react";

type BrandProps = {
  /** Veličina teksta logotipa (Chakra fontSize). */
  fontSize?: SystemStyleObject["fontSize"];
};

/**
 * Tekstualni logotip sistema.
 *
 * Po dizajn specifikaciji: „HR" je Inter **Bold / 700**, „Sistem" je
 * Inter **SemiBold / 600**. „HR" nosi tamni navy (#123B68), a „Sistem"
 * primarnu plavu (#1674D1) iz brend palete.
 */
export default function Brand({ fontSize = "xl" }: BrandProps) {
  return (
    <HStack as="span" gap={1.5} align="baseline" letterSpacing="-0.02em">
      <Text as="span" fontSize={fontSize} fontWeight={700} color="brand.800">
        HR
      </Text>
      <Text as="span" fontSize={fontSize} fontWeight={600} color="brand.600">
        Sistem
      </Text>
    </HStack>
  );
}
