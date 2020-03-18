#!/usr/bin/env bash -eux

# If we have virtualenv, activate it

if [[ -d ~/.py3 ]]; then
	VIRTUAL_ENV_DISABLE_PROMPT=true . ~/.py3/bin/activate
	python -V | egrep '^Python 3'   # fail on python 2
fi

# End of file