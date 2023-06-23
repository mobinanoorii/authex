FROM scratch
# Run the hello binary.
ENTRYPOINT [ "/authex" ]
CMD [ "server", "start" ]

