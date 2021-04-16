#!/bin/bash
if [[ -z "${GATEWAY}" ]]; then
	echo "No Gateway found, will not change routing"
else
	echo "changing default gateway for private subnets to ${GATEWAY}"

	ip route add 10.0.0.0/8 via ${GATEWAY}
	ip route add 172.16.0.0/12 via ${GATEWAY}
	ip route add 192.168.0.0/16 via ${GATEWAY}
	echo "changed routes for 10.0.0.0/8, 172.16.0.0/12 and 192.168.0.0/16"
	echo "public nets are still available"
fi
echo "replacing own process"
exec /app/main
