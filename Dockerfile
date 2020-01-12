FROM ubuntu

COPY main /online/exe/main
COPY templates /online/templates
COPY assets /online/assets
COPY DBSettings.json /online/DBSettings.json
COPY PushSettings.json /online/PushSettings.json

WORKDIR /online/exe/
ENTRYPOINT [ "./main" ]