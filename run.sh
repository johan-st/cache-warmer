#!/bin/bash

# This script is used to run the application
# check if gum is installed
if ! command -v gum &> /dev/null
then
    echo " - gum could not be found - "
    echo "Get gum here: https://github.com/charmbracelet/gum)"
    exit
fi

version=${"cache-warmer" --version}

COL_PRIME="#f80"
COL_BACK="#333"
MARGIN_FRAME="1 2"
PADDING_FRAME="2 4"
WIDTH="50"
BORDER="normal"
# SCRIPTS
# selection=$(gum choose "Run" "Build" "Build and Run" "Quit")


# --border-background "${COL_BACK}" \
gum style \
--align center \
--background "${COL_BACK}" \
--foreground "${COL_PRIME}" \
--border "${BORDER}" \
--border-foreground "${COL_PRIME}" \
--margin "${MARGIN_FRAME}" \
--padding "${PADDING_FRAME}" \
--width "${WIDTH}" \
"Cache Warmer" "${version}"
  # --border normal \
  # --border-foreground "#f80" \
  # --border-background "#333" \
  # --background "#333" \
	# --align center \
  # --margin "1 2" \
  # --padding "2 4" \

# targetUrl=$(gum input --placeholder "Target URL")