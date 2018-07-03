#!/bin/bash

# Maven projects
# clojure/clojure
# sendgrid/sendgrid-java
# google/truth
# google/auto
git clone https://github.com/apache/hadoop maven
cd maven
git checkout e65ff1c8be48ef4f04ed96f96ac4caef4974944d
cd ..

# Scala projects
# apache/spark
# linkerd/linkerd
# graphcool/prisma
# twitter/finagle
git clone https://github.com/graphcool/prisma sbt
cd sbt
git checkout 90058a94b096b3c0dece8df332288755632981de
cd ..