package theme

import "charm.land/huh/v2"

// Huh returns the huh form theme for this suite. Always pass it to
// huh.NewForm(...).WithTheme(theme.Huh()) so prompts match the rest of the UI.
func Huh() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		s := huh.ThemeBase(isDark)

		s.Focused.Base = s.Focused.Base.BorderForeground(Cyan)
		s.Focused.Title = s.Focused.Title.Foreground(Cyan).Bold(true)
		s.Focused.Description = s.Focused.Description.Foreground(Steel)
		s.Focused.ErrorIndicator = s.Focused.ErrorIndicator.Foreground(Crimson)
		s.Focused.ErrorMessage = s.Focused.ErrorMessage.Foreground(Crimson)
		s.Focused.SelectSelector = s.Focused.SelectSelector.Foreground(Magenta)
		s.Focused.SelectedOption = s.Focused.SelectedOption.Foreground(Cyan)
		s.Focused.FocusedButton = s.Focused.FocusedButton.Foreground(Ink).Background(Magenta).Bold(true)
		s.Focused.BlurredButton = s.Focused.BlurredButton.Foreground(Snow).Background(Night)
		s.Focused.TextInput.Cursor = s.Focused.TextInput.Cursor.Foreground(Magenta)
		s.Focused.TextInput.Prompt = s.Focused.TextInput.Prompt.Foreground(Cyan)
		s.Focused.TextInput.Placeholder = s.Focused.TextInput.Placeholder.Foreground(Slate)
		s.Focused.TextInput.Text = s.Focused.TextInput.Text.Foreground(Snow)

		s.Blurred = s.Focused
		s.Blurred.Base = s.Blurred.Base.BorderForeground(Slate)
		s.Blurred.Title = s.Blurred.Title.Foreground(Steel)
		s.Blurred.TextInput.Prompt = s.Blurred.TextInput.Prompt.Foreground(Slate)

		return s
	})
}
