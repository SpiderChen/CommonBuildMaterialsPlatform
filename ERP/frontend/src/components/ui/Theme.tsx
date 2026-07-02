import { createContext, type CSSProperties, type ReactNode, useContext, useMemo } from "react";
import { cx } from "./utils";

export type UiThemeTokenValue = string | number;

export type UiTheme = {
  colorPrimary?: UiThemeTokenValue;
  colorPrimaryContrast?: UiThemeTokenValue;
  colorSurface?: UiThemeTokenValue;
  colorSurfaceMuted?: UiThemeTokenValue;
  colorText?: UiThemeTokenValue;
  colorMuted?: UiThemeTokenValue;
  colorBorder?: UiThemeTokenValue;
  colorBorderSoft?: UiThemeTokenValue;
  colorDanger?: UiThemeTokenValue;
  colorSuccess?: UiThemeTokenValue;
  radiusSm?: UiThemeTokenValue;
  radiusMd?: UiThemeTokenValue;
  shadowDialog?: UiThemeTokenValue;
  dialogBackdrop?: UiThemeTokenValue;
  dialogZIndex?: UiThemeTokenValue;
  dialogSurface?: UiThemeTokenValue;
  dialogText?: UiThemeTokenValue;
  dialogMuted?: UiThemeTokenValue;
  dialogBorder?: UiThemeTokenValue;
  dialogBorderSoft?: UiThemeTokenValue;
  dialogRadius?: UiThemeTokenValue;
  dialogShadow?: UiThemeTokenValue;
  dialogMaxHeight?: UiThemeTokenValue;
  dialogMaxWidthSm?: UiThemeTokenValue;
  dialogMaxWidthMd?: UiThemeTokenValue;
  dialogMaxWidthLg?: UiThemeTokenValue;
  dialogMaxWidthXl?: UiThemeTokenValue;
  dialogMaxWidthWide?: UiThemeTokenValue;
  dialogHeaderPadding?: UiThemeTokenValue;
  dialogBodyPadding?: UiThemeTokenValue;
  dialogFooterPadding?: UiThemeTokenValue;
  dialogFeedbackSuccessBackground?: UiThemeTokenValue;
  dialogFeedbackSuccessBorder?: UiThemeTokenValue;
  dialogFeedbackErrorBackground?: UiThemeTokenValue;
  dialogFeedbackErrorBorder?: UiThemeTokenValue;
};

export type UiThemeStyle = CSSProperties & Record<`--${string}`, UiThemeTokenValue>;

const uiThemeTokens: Record<keyof UiTheme, `--${string}`> = {
  colorPrimary: "--ui-color-primary",
  colorPrimaryContrast: "--ui-color-primary-contrast",
  colorSurface: "--ui-color-surface",
  colorSurfaceMuted: "--ui-color-surface-muted",
  colorText: "--ui-color-text",
  colorMuted: "--ui-color-muted",
  colorBorder: "--ui-color-border",
  colorBorderSoft: "--ui-color-border-soft",
  colorDanger: "--ui-color-danger",
  colorSuccess: "--ui-color-success",
  radiusSm: "--ui-radius-sm",
  radiusMd: "--ui-radius-md",
  shadowDialog: "--ui-shadow-dialog",
  dialogBackdrop: "--ui-dialog-backdrop",
  dialogZIndex: "--ui-dialog-z-index",
  dialogSurface: "--ui-dialog-surface",
  dialogText: "--ui-dialog-text",
  dialogMuted: "--ui-dialog-muted",
  dialogBorder: "--ui-dialog-border",
  dialogBorderSoft: "--ui-dialog-border-soft",
  dialogRadius: "--ui-dialog-radius",
  dialogShadow: "--ui-dialog-shadow",
  dialogMaxHeight: "--ui-dialog-max-height",
  dialogMaxWidthSm: "--ui-dialog-max-width-sm",
  dialogMaxWidthMd: "--ui-dialog-max-width-md",
  dialogMaxWidthLg: "--ui-dialog-max-width-lg",
  dialogMaxWidthXl: "--ui-dialog-max-width-xl",
  dialogMaxWidthWide: "--ui-dialog-max-width-wide",
  dialogHeaderPadding: "--ui-dialog-header-padding",
  dialogBodyPadding: "--ui-dialog-body-padding",
  dialogFooterPadding: "--ui-dialog-footer-padding",
  dialogFeedbackSuccessBackground: "--ui-dialog-feedback-success-bg",
  dialogFeedbackSuccessBorder: "--ui-dialog-feedback-success-border",
  dialogFeedbackErrorBackground: "--ui-dialog-feedback-error-bg",
  dialogFeedbackErrorBorder: "--ui-dialog-feedback-error-border"
};

const UiThemeContext = createContext<UiTheme>({});

export function createUiThemeStyle(theme: UiTheme = {}) {
  const style = {} as UiThemeStyle;
  (Object.keys(uiThemeTokens) as Array<keyof UiTheme>).forEach((token) => {
    const value = theme[token];
    if (value !== undefined) {
      style[uiThemeTokens[token]] = value;
    }
  });
  return style;
}

export type UiThemeProviderProps = {
  children: ReactNode;
  className?: string;
  style?: CSSProperties;
  theme?: UiTheme;
};

export function UiThemeProvider({ children, className, style, theme = {} }: UiThemeProviderProps) {
  const themeStyle = useMemo(() => createUiThemeStyle(theme), [theme]);

  return (
    <UiThemeContext.Provider value={theme}>
      <div className={cx("ui-theme", className)} style={{ ...themeStyle, ...style }}>
        {children}
      </div>
    </UiThemeContext.Provider>
  );
}

export function useUiTheme() {
  return useContext(UiThemeContext);
}
