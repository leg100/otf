helm lint ./charts/otfd
helm lint ./charts/otf-agent
helm template ./charts/otfd > /dev/null
helm template ./charts/otfd --values ./charts/otfd/tests/volume_mounts.yaml > /dev/null
helm template ./charts/otf-agent --set token=my_agent_token --set url=https://otf.ninja > /dev/null
