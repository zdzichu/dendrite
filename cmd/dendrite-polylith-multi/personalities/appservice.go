// Copyright 2020 The Matrix.org Foundation C.I.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package personalities

import (
	"github.com/matrix-org/dendrite/appservice"
	"github.com/matrix-org/dendrite/setup/base"
	basepkg "github.com/matrix-org/dendrite/setup/base"
	"github.com/matrix-org/dendrite/setup/config"
)

func Appservice(base *base.BaseDendrite, cfg *config.Dendrite) {
	userAPI := base.UserAPIClient()
	rsAPI := base.RoomserverHTTPClient()

	intAPI := appservice.NewInternalAPI(base, userAPI, rsAPI)
	appservice.AddInternalRoutes(base.InternalAPIMux, intAPI, base.EnableMetrics)

	base.SetupAndServeHTTP(
		base.Cfg.AppServiceAPI.InternalAPI.Listen, // internal listener
		basepkg.NoListener,                        // external listener
		nil, nil,
	)
}
